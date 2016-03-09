import logging

logger = logging.getLogger()
logger.setLevel(logging.NOTSET)


def run(event, context):
    logger.info('an info string')
    logger.warning('a warning string')
    logger.debug('a debug string')
    logger.error('something went wrong')
    logger.critical('end of game')
    
# logger.exception produces different result due to different test.py file location on AWS Lambda server and in the test's docker container
#    try:
#       1/0
#    except Exception as e:
#        logger.exception("rabbit eats a fox")
