# -*- encoding: utf-8 -*-

import os
import urlparse
import requests
import datetime
import microdata
from mf2py.parser import Parser
from flask import Flask, session, jsonify, redirect, request, render_template, abort, redirect

import redis as r
url = urlparse.urlparse(os.environ.get('REDISCLOUD_URL'))
redis = r.Redis(host=url.hostname, port=url.port, password=url.password)

app = Flask(__name__)
app.debug = True

@app.route('/<path:addr>/')
def avatar(addr):
    if not addr.startswith('https://') and not addr.startswith('http://'):
        addr = 'http://' + addr
    addr = urlparse.urlparse(addr)
    host = addr.netloc
    path = addr.path if addr.path and addr.path != '/' else ''

    # search in our cache
    cached_src = redis.get(host)
    if not cached_src and path:
        cached_src = redis.get(host + path)
    if cached_src:
        return redirect(cached_src)

    else:
        # otherwise fetch it from the live page
        protocol = addr.scheme

        # try microformats2
        src = None
        base = protocol + '://' + host + path
        try:
            parsed = Parser(url=base).to_dict()
            # try rel=icon
            if 'icon' in parsed['rels']:
                src = parsed['rels']['icon'][-1]
            # try h-card photo
            if not src:
                for item in parsed['items']:
                    if u'h-card' in item['type']:
                        if u'photo' in item['properties']:
                            src = item['properties']['photo'][0]
                            break

        except requests.exceptions.ConnectionError:
            pass

        # try microdata
        if not src:
            try:
                html = requests.get(base).text.encode('utf-8')
                items = microdata.get_items(html)
                if len(items):
                    image = items[0].image
                    if image:
                        src = image if image.startswith('http://') or image.startswith('https://') else urlparse.urljoin(base, image)
            except requests.exceptions.ConnectionError:
                pass

        # when we find nothing
        if not src:
            src = 'http://robohash.org/' + host + path

        # save to redis
        redis.setex(host + path, src, datetime.timedelta(days=15))

        # return
        return redirect(src)
