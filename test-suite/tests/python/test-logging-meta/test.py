import logging


def run(event, context):
    logger = logging.getLogger()
    print ("logger.name = ", logger.name)
    print ("logger.level = ", logger.level)

