FROM python:3.14-slim

RUN apt-get update && apt-get install -y --no-install-recommends tini \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

RUN groupadd -r appuser && useradd -r -g appuser -d /app appuser && chown -R appuser:appuser /app

COPY exporter.py config_loader.py shelly_server.py entrypoint.sh ./
COPY collectors/ ./collectors/

RUN chmod +x entrypoint.sh

USER appuser

EXPOSE 9102 9101 8765 8766

ENTRYPOINT ["tini", "--", "./entrypoint.sh"]
