#!/usr/bin/env python3
import logging
import os

def main():
    initialize_log('DEBUG')
    os.system('echo -n TEXT | nc -N server 12345 > out.txt')
    f = open("out.txt", "r")
    textRead = f.read()
    if ("TEXT" in textRead):
        logging.debug("Server is working correctly")
    else:
        logging.debug("Server is not working")

def initialize_log(logging_level):
    """
    Python custom logging initialization

    Current timestamp is added to be able to identify in docker
    compose logs the date when the log has arrived
    """
    logging.basicConfig(
        format='%(asctime)s %(levelname)-8s %(message)s',
        level=logging_level,
        datefmt='%Y-%m-%d %H:%M:%S',
    )


if __name__ == "__main__":
    main()
