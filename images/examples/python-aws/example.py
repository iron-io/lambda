import math
import random
import boto3

from wand.image import Image
from io import BytesIO


def run(event,context):
    #CHANGE THE BUCKET NAME AND THE KEY TO YOUR OWN
    bucketName = event['bucket']
    imageKey = event['key']

    s3 = boto3.resource('s3')

    imageObject = s3.Object(bucket_name=bucketName, key=imageKey)

    imageStream = imageObject.get()['Body']

    with Image(file=imageStream) as image:
        for x in range(0,10):
            flipflop(image)
        result = BytesIO()
        image.save(file=result)
        result.seek(0)

    imageObject.put(Body=result, ACL='public-read', ContentType='image/jpeg',)


def flipflop(image):
    xblock, xcount = selectBlockSize(image.width)
    yblock, ycount = selectBlockSize(image.height)

    x0, y0, x1, y1 = 0, 0, 0, 0
    while x0 >= x1 or y0 >= y1:
        x0 = int(math.floor(random.random() * xcount + 0) * xblock)
        x1 = int(math.floor(random.random() * xcount + 1) * xblock)
        y0 = int(math.floor(random.random() * ycount + 0) * yblock)
        y1 = int(math.floor(random.random() * ycount + 1) * yblock)
        if x1 < x0:
            x0, x1 = x1, x0
        if y1 < y0:
            y0, y1 = y1, y0

    with image.clone() as copy:
        copy.crop(left=x0, top=y0, right=x1, bottom=y1)
        if random.random() < 0.5:
            copy.flip()
        else:
            copy.flop()
        image.composite(image=copy, left=x0, top=y0)


def selectBlockSize(size):
    for count in [16, 12, 8, 20, 15]:
        if size % count == 0:
            break
    count = count or 8
    block = math.floor(size / count)
    return block, count
