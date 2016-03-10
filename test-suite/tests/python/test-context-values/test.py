from __future__ import print_function
import time


def run(event, context):
    print("Function name is set:", not (context.function_name is None))
    print("Function version is set:", not (context.function_version is None))
    print("Memory limit in MB (grater or equal than 100):", context.memory_limit_in_mb > 100)
    print("AWS request ID is set:", not (context.aws_request_id is None))

    remaining1 = context.get_remaining_time_in_millis()
    time.sleep(1)
    remaining2 = context.get_remaining_time_in_millis()
    remainingDelta = remaining1 - remaining2
    print("Time remaining delta (MS) is around 1 second:",
        remainingDelta >= 900 and remainingDelta <= 1100)

#    print("Log group name:", context.log_group_name)
#    print("Log stream name:", context.log_stream_name)
#    print("Invoked function ARN:", context.invoked_function_arn)
#    print("Identity:", context.identity)
#    print("Client context:", context.client_context)
