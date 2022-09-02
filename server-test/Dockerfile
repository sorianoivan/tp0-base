FROM python:3.9.7-slim
RUN apt-get update && apt-get install -y netcat
COPY test.py /
ENTRYPOINT ["/bin/sh"]