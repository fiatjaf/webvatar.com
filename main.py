# -*- encoding: utf-8 -*-

from __future__ import division

import os
import math
import urlparse
import requests
import datetime
import hashlib
import microdata
from mf2py.parser import Parser
from flask import Flask, session, jsonify, redirect, request, render_template, abort, redirect

import redis as r
url = urlparse.urlparse(os.environ.get('REDISCLOUD_URL'))
redis = r.Redis(host=url.hostname, port=url.port, password=url.password)

app = Flask(__name__)
app.debug = True

@app.route('/')
def index():
    return redirect('http://indiewebcamp.com/webvatar')

@app.route('/<path:addr>/')
def avatar(addr):
    if not addr.startswith('https://') and not addr.startswith('http://'):
        addr = 'http://' + addr
    addr = urlparse.urlparse(addr)
    host = addr.netloc
    path = addr.path if addr.path and addr.path != '/' else ''

    # search in our cache
    cached = redis.get(host + path) or redis.get(host) if path else None

    if cached:
        # if something is found, send it
        return redirect(cached)

    elif cached == None:
        # otherwise try to fetch from the live page
        protocol = addr.scheme

        base = protocol + '://' + host + path
        try:
            res = requests.get(base, verify=False)
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
                alt.consider(src)
        # try h-card photo
        for item in parsed['items']:
            if 'h-card' in item['type']:
                if 'photo' in item['properties']:
                    for src in item['properties']['photo']:
                        alt.consider(src)

        # try microdata
        items = microdata.get_items(html)
        if len(items):
            for item in items:
                alt.consider(item.image)

        # get best (or blank string)
        ignoresmall = not request.args.get('acceptsmall', False)
        final = alt.best(ignoresmall=ignoresmall)

        # save to redis
        redis.setex(host + path, final, datetime.timedelta(days=3))

        # return
        if final:
            return redirect(final)

    # if nothing was found until now, send an alt
    if request.args.get('alt') == 'robohash':
        return redirect(alt.robohash())
    elif request.args.get('alt') == 'hash':
        return redirect(alt.hashshow())
    else:
        return redirect(alt.nameshow())

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

    def best(self, ignoresmall=True):
        eligible = self.considering
        if ignoresmall:
            eligible = filter(lambda x: x['size'] > 4500, eligible)
        if len(eligible) == 0:
            return ''
        ordered = sorted(eligible, key=lambda x: x['size'], reverse=True)
        return ordered[0]['url']

    def complete(self, url):
        if url.startswith('http://') or url.startswith('https://'):
            return url
        else:
            return urlparse.urljoin(self.base, url)

    def robohash(self):
        return 'http://robohash.org/' + self.host + self.path

    def hashshow(self):
        l = hashlib.sha256(self.host + self.path).hexdigest()
        lines = '|'.join((l[0:11], l[11:21], l[21:32], l[32:43], l[43:52], l[52:64]))
        url = 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|32|r|000|_|' + lines
        return url

    def nameshow(self):
        line = self.host + self.path
        linelen = len(line)
        n = linelen/3 if linelen > 30 else linelen/2 if linelen > 20 else linelen
        n = int(math.ceil(n))
        lines = '|'.join([line[i:i+n] for i in range(0, linelen, n)])
        url = 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||' + lines + '||'
        return url
