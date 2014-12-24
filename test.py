import sys
import requests

expectations = (
    ('/tantek.com', 'http://tantek.com/photo.jpg'),
    ('/werd.io/', 'http://werd.io/file/538d0a4cbed7de5111a1ad31/thumb.jpg'),
    ('/aaronparecki.com', 'http://aaronparecki.com/images/aaronpk-256.jpg'),
    ('/snarfed.org', 'http://snarfed.org/ryan_profile_square_thumb.jpg'),
    ('/ben.thatmustbe.me/?alt=robohash', 'https://ben.thatmustbe.me/image/static/icon_144.jpg'),
    ('/kylewm.com/', 'https://kylewm.com/static/img/users/kyle.jpg'),
    ('/stream.withknown.com', 'http://stream.withknown.com/file/537431cfbed7de93520ba75d/thumb.jpg'),
    ('/fiatjaf.alhur.es/?alt=robohash', 'http://robohash.org/fiatjaf.alhur.es'),
    ('/cweiske.de', 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||cweiske.de||'),
    ('/cweiske.de?acceptsmall=true', 'http://cweiske.de/favicon.ico'),
)

try:
    base = sys.argv[1]
except:
    base = 'http://0.0.0.0:5000'

errors = []
successes = []
for url, expected in expectations:
    try:
        location = requests.head(base + url).headers['Location']
    except:
        location = '{ request error }'
    print ' - {}: expected {}\n     {}got {}'.format(url, expected, len(url)*' ', location)
    if location != expected:
        errors.append(url)
    else:
        successes.append(url)
    
print
print 'ERRORS'
for error in errors:
    print ' -', error
print
print 'SUCCESSES'
for success in successes:
    print ' -', success
