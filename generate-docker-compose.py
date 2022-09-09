import sys
import argparse

SERVER_CONFIG = """version: '3'
services:
  server:
    container_name: server
    image: server:latest
    volumes:
      - ./server/config.ini:/config.ini
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_PORT=12345
      - SERVER_LISTEN_BACKLOG=7
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
      
"""

CLIENT_CONFIG = """  client{}:
    container_name: client{}
    image: client:latest
    volumes:
      - ./client/config.yaml:/config.yaml
    entrypoint: /client
    environment:
      - CLI_ID={}
      - CLI_SERVER_ADDRESS=server:12345
      - CLI_LOOP_LAPSE=1m2s
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

"""

NETWORK_CONFIG = """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

DOCKER_COMPOSE_FILENAME = "docker-compose-dev.yaml"

def main(clients_amount):
    f = open(DOCKER_COMPOSE_FILENAME, "w")
    f.write(SERVER_CONFIG)
    for i in range(1,clients_amount + 1):
        f.write(CLIENT_CONFIG.format(i,i,i))
    f.write(NETWORK_CONFIG)
    f.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Generate docker-compose file with N clients')
    parser.add_argument('num_clients', metavar='N', type=int,
                    help='The number of clients')
    args = parser.parse_args()
    main(args.num_clients)