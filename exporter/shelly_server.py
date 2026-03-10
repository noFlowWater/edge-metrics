#!/usr/bin/env python3
"""
Shelly WebSocket Server
각 디바이스의 쉘리플러그가 연결하여 실시간 전력 데이터를 전송하는 독립 서버
"""
import asyncio
import json
import logging
import os
import time
from threading import Lock
from aiohttp import web
import websockets


class ShellyConnectionRegistry:
    """
    WebSocket 연결 레지스트리 (Thread-safe)
    각 디바이스의 WebSocket 연결을 추적
    """

    def __init__(self):
        self.connections = {}  # {device_id: websocket}
        self.lock = Lock()
        self.logger = logging.getLogger("ShellyConnectionRegistry")

    def register(self, device_id: str, websocket):
        """연결 등록 (thread-safe)"""
        with self.lock:
            self.connections[device_id] = websocket
            self.logger.info(f"Registered device: {device_id}")

    def unregister(self, device_id: str):
        """연결 해제 (thread-safe)"""
        with self.lock:
            if device_id in self.connections:
                del self.connections[device_id]
                self.logger.info(f"Unregistered device: {device_id}")

    def get_connection(self, device_id: str):
        """WebSocket 연결 조회 (thread-safe)"""
        with self.lock:
            return self.connections.get(device_id)

    def get_all_devices(self) -> list:
        """디바이스 목록 조회 (thread-safe)"""
        with self.lock:
            return list(self.connections.keys())


class ShellyWebSocketHandler:
    """
    WebSocket 연결 처리
    Shelly 플러그의 Outbound WebSocket 연결을 처리하고 RPC 요청/응답 처리
    """

    def __init__(self, registry: ShellyConnectionRegistry):
        self.registry = registry
        self.pending_requests = {}  # {request_id: asyncio.Future}
        self.logger = logging.getLogger("ShellyWebSocketHandler")

    async def handle_connection(self, websocket):
        """
        WebSocket 연결 처리 메인 루프

        Args:
            websocket: WebSocket 연결 객체
        """
        device_id = None
        remote_addr = websocket.remote_address

        try:
            self.logger.info(f"New WebSocket connection from {remote_addr}")

            # 메시지 수신 루프
            async for message in websocket:
                try:
                    # JSON-RPC 메시지 파싱
                    data = self._parse_rpc_message(message)

                    if not data:
                        continue

                    # 디바이스 ID 추출 및 등록 (첫 메시지에서)
                    if not device_id:
                        device_id = self._extract_device_id(data, remote_addr)
                        if device_id:
                            self.registry.register(device_id, websocket)
                            self.logger.info(f"Device identified and registered: {device_id}")

                    # RPC 응답 디스패치
                    if self.dispatch_rpc_response(data):
                        continue

                    # Push notifications 무시 (NotifyStatus, NotifyFullStatus, etc.)

                except Exception as e:
                    self.logger.error(f"Error processing message: {e}")
                    continue

        except websockets.exceptions.ConnectionClosed:
            self.logger.info(f"Connection closed: {remote_addr}")
        except Exception as e:
            self.logger.error(f"WebSocket error: {e}")
        finally:
            # 연결 종료 시 디바이스 등록 해제
            if device_id:
                self.registry.unregister(device_id)

    def _parse_rpc_message(self, message: str) -> dict:
        """
        JSON-RPC 메시지 파싱

        Args:
            message: 원시 JSON 문자열

        Returns:
            파싱된 딕셔너리 또는 None
        """
        try:
            return json.loads(message)
        except Exception as e:
            self.logger.error(f"Failed to parse RPC message: {e}")
            return None

    def _extract_device_id(self, data: dict, remote_addr: tuple) -> str:
        """
        RPC 메시지에서 디바이스 ID 추출

        Args:
            data: 파싱된 RPC 메시지
            remote_addr: 원격 주소 (IP, port)

        Returns:
            디바이스 ID 문자열
        """
        # Try to get device ID from message "src" field
        if "src" in data:
            return data["src"]

        # Fallback: use remote IP address
        return f"shelly_{remote_addr[0].replace('.', '_')}"

    async def send_rpc_request(self, websocket, method: str, params: dict = None):
        """
        RPC 요청 전송 및 응답 대기

        Args:
            websocket: WebSocket 연결 객체
            method: RPC 메서드 이름 (예: "Switch.GetStatus")
            params: RPC 파라미터 (예: {"id": 0})

        Returns:
            RPC 응답 메시지

        Raises:
            Exception: RPC 타임아웃 또는 전송 실패
        """
        import uuid

        request_id = str(uuid.uuid4())
        request = {
            "id": request_id,
            "method": method,
            "params": params or {}
        }

        # Create Future for response
        future = asyncio.Future()
        self.pending_requests[request_id] = future

        try:
            # Send RPC request
            await websocket.send(json.dumps(request))
            self.logger.debug(f"Sent RPC request: {method} (id: {request_id})")

            # Wait for response with 5s timeout
            response = await asyncio.wait_for(future, timeout=5.0)
            self.logger.debug(f"Received RPC response (id: {request_id})")
            return response

        except asyncio.TimeoutError:
            self.logger.error(f"RPC request timeout: {method} (id: {request_id})")
            raise Exception("RPC request timeout")
        finally:
            self.pending_requests.pop(request_id, None)

    def dispatch_rpc_response(self, message: dict) -> bool:
        """
        RPC 응답을 대기 중인 요청으로 라우팅

        Args:
            message: 수신된 RPC 메시지

        Returns:
            True if message was an RPC response, False otherwise
        """
        if "id" in message and message["id"] in self.pending_requests:
            future = self.pending_requests[message["id"]]
            if not future.done():
                future.set_result(message)
                self.logger.debug(f"Dispatched RPC response (id: {message['id']})")
            return True
        return False

    def _extract_metrics_from_rpc_result(self, result: dict) -> dict:
        """
        Switch.GetStatus RPC 결과에서 모든 메트릭 추출

        Args:
            result: RPC response의 "result" 필드

        Returns:
            메트릭 딕셔너리 (모든 Shelly Switch 메트릭 포함, shelly_ prefix)
        """
        metrics = {}

        try:
            # Switch output state (boolean → 1/0)
            # output: true if the output channel is currently on, false otherwise
            if "output" in result:
                metrics["shelly_power_switch_output"] = 1 if result["output"] else 0

            # Instantaneous active power (Watts)
            # apower: Last measured instantaneous active power delivered to the attached load
            if "apower" in result:
                metrics["shelly_power_total_watts"] = float(result["apower"])

            # Supply voltage (Volts)
            # voltage: Last measured voltage
            if "voltage" in result:
                metrics["shelly_power_voltage_volts"] = float(result["voltage"])

            # Current (Amperes)
            # current: Last measured current
            if "current" in result:
                metrics["shelly_power_current_amps"] = float(result["current"])

            # Power factor
            # pf: Last measured power factor
            if "pf" in result:
                metrics["shelly_power_factor"] = float(result["pf"])

            # Network frequency (Hz)
            # freq: Last measured network frequency
            if "freq" in result:
                metrics["shelly_power_frequency_hz"] = float(result["freq"])

            # Active energy counter (Watt-hours)
            # aenergy: Information about the active energy counter
            if "aenergy" in result and isinstance(result["aenergy"], dict):
                aenergy = result["aenergy"]

                # Total consumed energy (Watt-hours)
                # total: Total energy consumed
                if "total" in aenergy:
                    metrics["shelly_energy_total_wh"] = float(aenergy["total"])

                # Energy by minute (Milliwatt-hours for last 3 complete minutes)
                # by_minute: Total energy flow for the last three complete minutes
                # The 0-th element indicates counts during the minute preceding minute_ts
                if "by_minute" in aenergy and isinstance(aenergy["by_minute"], list):
                    for i, value in enumerate(aenergy["by_minute"][:3]):
                        metrics[f"shelly_energy_minute_{i}_mwh"] = float(value)

                # Timestamp of current minute start (Unix timestamp in UTC)
                # minute_ts: Unix timestamp marking the start of the current minute
                if "minute_ts" in aenergy:
                    metrics["shelly_energy_minute_timestamp"] = int(aenergy["minute_ts"])

            # Returned active energy counter (Watt-hours)
            # ret_aenergy: Information about the returned active energy counter
            # Note: Returned energy is also added to aenergy container
            if "ret_aenergy" in result and isinstance(result["ret_aenergy"], dict):
                ret_aenergy = result["ret_aenergy"]

                # Total returned energy (Watt-hours)
                # total: Total returned energy consumed
                if "total" in ret_aenergy:
                    metrics["shelly_energy_returned_total_wh"] = float(ret_aenergy["total"])

                # Returned energy by minute (Milliwatt-hours for last 3 complete minutes)
                # by_minute: Returned energy for the last three complete minutes
                if "by_minute" in ret_aenergy and isinstance(ret_aenergy["by_minute"], list):
                    for i, value in enumerate(ret_aenergy["by_minute"][:3]):
                        metrics[f"shelly_energy_returned_minute_{i}_mwh"] = float(value)

                # Timestamp of current minute start (Unix timestamp in UTC)
                # minute_ts: Unix timestamp marking the start of the current minute
                if "minute_ts" in ret_aenergy:
                    metrics["shelly_energy_returned_minute_timestamp"] = int(ret_aenergy["minute_ts"])

            # Temperature measurements
            # temperature: Information about the temperature (if applicable)
            if "temperature" in result and isinstance(result["temperature"], dict):
                temp = result["temperature"]

                # Temperature in Celsius (null if out of measurement range)
                # tC: Temperature in Celsius
                if "tC" in temp and temp["tC"] is not None:
                    metrics["shelly_temperature_celsius"] = float(temp["tC"])

                # Temperature in Fahrenheit (null if out of measurement range)
                # tF: Temperature in Fahrenheit
                if "tF" in temp and temp["tF"] is not None:
                    metrics["shelly_temperature_fahrenheit"] = float(temp["tF"])

            # Error conditions
            # errors: Error conditions occurred (overtemp, overpower, overvoltage, undervoltage)
            if "errors" in result and isinstance(result["errors"], list):
                # Number of active errors
                metrics["shelly_errors_count"] = len(result["errors"])

                # Individual error flags
                error_types = ["overtemp", "overpower", "overvoltage", "undervoltage"]
                for error_type in error_types:
                    metrics[f"shelly_error_{error_type}"] = 1 if error_type in result["errors"] else 0

        except Exception as e:
            self.logger.error(f"Error extracting metrics from RPC result: {e}")

        return metrics


class ShellyHTTPHandler:
    """
    HTTP API 처리
    ShellyCollector가 메트릭을 조회할 수 있는 HTTP API 제공 (RPC 방식)
    """

    def __init__(self, registry: ShellyConnectionRegistry, ws_handler: ShellyWebSocketHandler):
        self.registry = registry
        self.ws_handler = ws_handler
        self.logger = logging.getLogger("ShellyHTTPHandler")

    async def handle_metrics(self, request):
        """
        GET /metrics
        현재 연결된 유일한 디바이스의 메트릭 반환 (On-demand RPC 방식)

        동작:
        - Shelly plug 연결 전: 404 (No Shelly device connected)
        - Shelly plug 연결 후: RPC로 실시간 메트릭 조회

        Returns:
            JSON 응답
        """
        try:
            devices = self.registry.get_all_devices()

            if not devices:
                return web.json_response(
                    {"error": "No Shelly device connected"},
                    status=404
                )

            # 첫 번째 (유일한) 디바이스 선택
            device_id = devices[0]
            websocket = self.registry.get_connection(device_id)

            if not websocket:
                return web.json_response(
                    {"error": "Device connection lost"},
                    status=502
                )

            # RPC 요청 전송: Switch.GetStatus
            response = await self.ws_handler.send_rpc_request(
                websocket,
                method="Switch.GetStatus",
                params={"id": 0}
            )

            # RPC 응답에서 메트릭 추출
            if "result" in response:
                metrics = self.ws_handler._extract_metrics_from_rpc_result(response["result"])

                return web.json_response({
                    "device_id": device_id,
                    "metrics": metrics,
                    "timestamp": time.time()
                })
            else:
                raise Exception("Invalid RPC response")

        except asyncio.TimeoutError:
            return web.json_response(
                {"error": "RPC request timeout"},
                status=504
            )
        except Exception as e:
            self.logger.error(f"Error getting metrics: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )

    async def handle_devices(self, request):
        """
        GET /devices
        연결된 디바이스 목록 반환

        Returns:
            JSON 응답
        """
        devices = self.registry.get_all_devices()

        return web.json_response({
            "devices": devices,
            "count": len(devices)
        })


class ShellyServer:
    """
    메인 서버
    WebSocket 서버와 HTTP API 서버를 동시에 실행
    """

    def __init__(self, ws_port=8765, http_port=8766):
        self.ws_port = ws_port
        self.http_port = http_port

        # 공유 연결 레지스트리
        self.registry = ShellyConnectionRegistry()

        # 핸들러 초기화
        self.ws_handler = ShellyWebSocketHandler(self.registry)
        self.http_handler = ShellyHTTPHandler(self.registry, self.ws_handler)

        self.logger = logging.getLogger("ShellyServer")

    async def start_websocket_server(self):
        """WebSocket 서버 시작"""
        try:
            async with websockets.serve(
                self.ws_handler.handle_connection,
                "0.0.0.0",
                self.ws_port,
                ping_interval=30,  # Send ping every 30s
                ping_timeout=10    # Close if no pong after 10s
            ):
                self.logger.info(f"🔌 WebSocket server started on port {self.ws_port}")
                await asyncio.Future()  # Run forever
        except Exception as e:
            self.logger.error(f"❌ WebSocket server failed: {e}")

    async def start_http_server(self):
        """HTTP API 서버 시작"""
        try:
            app = web.Application()

            # 라우트 등록
            app.router.add_get("/metrics", self.http_handler.handle_metrics)
            app.router.add_get("/devices", self.http_handler.handle_devices)

            runner = web.AppRunner(app)
            await runner.setup()

            site = web.TCPSite(runner, "0.0.0.0", self.http_port)
            await site.start()

            self.logger.info(f"🌐 HTTP API server started on port {self.http_port}")
            self.logger.info(f"   - GET :{self.http_port}/metrics")
            self.logger.info(f"   - GET :{self.http_port}/devices")

            await asyncio.Future()  # Run forever
        except Exception as e:
            self.logger.error(f"❌ HTTP server failed: {e}")

    def run(self):
        """서버 시작 (WebSocket + HTTP 동시 실행)"""
        asyncio.run(self._run_both())

    async def _run_both(self):
        """WebSocket과 HTTP 서버를 asyncio.gather로 동시 실행"""
        self.logger.info("🚀 Shelly Server starting...")

        await asyncio.gather(
            self.start_websocket_server(),
            self.start_http_server()
        )


def main():
    """메인 엔트리 포인트"""
    # 로깅 설정
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    logger = logging.getLogger(__name__)

    # 환경변수에서 포트 읽기
    ws_port = int(os.getenv("SHELLY_WS_PORT", "8765"))
    http_port = int(os.getenv("SHELLY_HTTP_PORT", "8766"))

    logger.info(f"Configuration:")
    logger.info(f"  WebSocket Port: {ws_port}")
    logger.info(f"  HTTP API Port: {http_port}")

    # 서버 시작
    server = ShellyServer(ws_port=ws_port, http_port=http_port)

    try:
        server.run()
    except KeyboardInterrupt:
        logger.info("Server stopped by user")
    except Exception as e:
        logger.error(f"Fatal error: {e}", exc_info=True)
        exit(1)


if __name__ == "__main__":
    main()
