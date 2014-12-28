# -*- encoding: utf-8 -*-

import os
import urlparse
import redis as r
url = urlparse.urlparse(os.environ.get('REDISCLOUD_URL'))
redis = r.Redis(host=url.hostname, port=url.port, password=url.password)

for k in redis.keys():
    print redis.delete(k)
