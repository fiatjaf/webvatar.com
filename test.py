import sys
import requests

expectations = (
    ('/tantek.com', 'http://tantek.com/photo.jpg'),
    ('/werd.io/', 'http://werd.io/file/538d0a4cbed7de5111a1ad31/thumb.jpg'),
    ('/aaronparecki.com', 'http://aaronparecki.com/images/aaronpk-256.jpg'),
    ('/snarfed.org', 'https://snarfed.org/ryan_profile_square_thumb.jpg'),
    ('/ben.thatmustbe.me/?alt=robohash', 'https://ben.thatmustbe.me/image/static/icon_144.jpg'),
    ('/kylewm.com/', 'https://kylewm.com/static/img/users/kyle.jpg'),
    ('/stream.withknown.com', 'http://stream.withknown.com/file/537431cfbed7de93520ba75d/thumb.jpg'),
    ('/fiatjaf.alhur.es/?alt=robohash', 'http://robohash.org/fiatjaf.alhur.es'),
    ('/cweiske.de?acceptsmall=n', 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||cweiske.de||'),
    ('/cweiske.de', 'http://cweiske.de/favicon.ico'),
    ('/fiatjaf.withknown.com', 'http://fiatjaf.withknown.com/file/5145a8fbe4ad4422adba1bf0e896d2a9/thumb.jpg'),
    ('/http://fiatjaf.withknown.com', 'http://fiatjaf.withknown.com/file/5145a8fbe4ad4422adba1bf0e896d2a9/thumb.jpg'),
    ('/fiatjaf.gmail.com.questo.email', 'https://secure.gravatar.com/avatar/b760f503c84d1bf47322f401066c753f?d=blank'),
    ('/http://fiatjaf.gmail.com.questo.email', 'https://secure.gravatar.com/avatar/b760f503c84d1bf47322f401066c753f?d=blank'),
    ('/ben.thatmustbe.me/?d=robohash&f=y', 'http://robohash.org/ben.thatmustbe.me')
    ('/ben.thatmustbe.me/?d=identicon&f=y', 'https://secure.gravatar.com/avatar/eb8de282b4131c4ab253871ff867c87a?d=identicon')
    ('/ben.thatmustbe.me/?d=identicon&f=y', 'https://secure.gravatar.com/avatar/eb8de282b4131c4ab253871ff867c87a?d=identicon')
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
