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
    #cached_src = redis.get(host)
    #if not cached_src and path:
    #    cached_src = redis.get(host + path)
    #if cached_src:
    #    return redirect(cached_src)

    if True:
        # otherwise fetch it from the live page
        protocol = addr.scheme

        base = protocol + '://' + host + path
        try:
            res = requests.get(base)
        except requests.exceptions.ConnectionError:
            return abort(404)
        base = res.url
        html = res.text.encode('utf-8')

        # alternatives to consider
        alt = Alternatives(base, host, path)

        # try microformats2
        parsed = Parser(doc=html).to_dict()
        # try rel=icon
        if 'icon' in parsed['rels']:
            for src in parsed['rels']['icon']:
                print src
                alt.consider(src)
        # try h-card photo
        for item in parsed['items']:
            if u'h-card' in item['type']:
                if u'photo' in item['properties']:
                    for src in item['properties']['photo']:
                        print src
                        alt.consider(src)

        # try microdata
        items = microdata.get_items(html)
        if len(items):
            for item in items:
                alt.consider(item.image)

        # use best
        final = alt.best() or alt.robohash()

        # save to redis
        #redis.setex(host + path, src, datetime.timedelta(days=15))

        # return
        return redirect(final)

class Alternatives(object):
    def __init__(self, base, host, path):
        self.considering = []
        self.host = host
        self.path = path
        self.base = base

    def consider(self, url):
        if not url or type(url) is not str and type(url) is not unicode:
            return

        url = self.complete(url)
        r = requests.head(url, verify=False)
        if not r.ok:
            return

        alternative = {'url': url, 'size': 6000}
        if 'content-length' in r.headers:
            alternative['size'] = int(r.headers['content-length'])
        self.considering.append(alternative)

    def best(self):
        if not self.considering:
            return None

        ordered = sorted(self.considering, key=lambda x: x['size'], reverse=True)
        print ordered
        return ordered[0]['url']

    def complete(self, url):
        if url.startswith('http://') or url.startswith('https://'):
            return url
        else:
            return urlparse.urljoin(self.base, url)

    def robohash(self):
        return 'http://robohash.org/' + self.host + self.path
