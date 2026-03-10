FROM python:3.12-alpine AS builder
WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

FROM python:3.12-alpine
LABEL org.opencontainers.image.source="https://github.com/noFlowWater/edge-metrics"

RUN apk add --no-cache tini && \
    rm -rf /usr/local/lib/python3.12/ensurepip \
           /usr/local/lib/python3.12/idlelib \
           /usr/local/bin/pip* \
           /usr/local/lib/python3.12/site-packages/pip*

WORKDIR /app

COPY --from=builder /install /usr/local

RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser -h /app appuser && chown -R appuser:appuser /app

COPY exporter.py config_loader.py shelly_server.py entrypoint.sh ./
COPY collectors/ ./collectors/

RUN chmod +x entrypoint.sh

USER appuser

EXPOSE 9102 9101 8765 8766

ENTRYPOINT ["tini", "--", "./entrypoint.sh"]
