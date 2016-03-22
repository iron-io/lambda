from __future__ import print_function
import boto3
import uuid


def run(event, context):
    print ("Test start")
    s3 = boto3.resource('s3')

    bucket = s3.Bucket(Bucket=event['bucket'])

    if bucket.creation_date is None:
        print ("Creating bucket ...")
        bucket.create(
            ACL='private',
            CreateBucketConfiguration={'LocationConstraint': event['region']}
            )
    bucket.wait_until_exists()

    print ("Adding object ...")
    obj = bucket.put_object(
        Key='myKey-'+str(uuid.uuid4()),
        Body='Hello!')

    print ("Deleting object ...")
    obj.delete()
    print ("Accessing S3 has been succeeded")
