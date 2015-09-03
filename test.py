import sys
import requests

expectations = (
    ('/tantek.com', 'http://tantek.com/photo.jpg'),
    ('/werd.io/', 'http://werd.io/gfx/logos/logo_k.png'),
    ('/aaronparecki.com', 'http://aaronparecki.com/images/aaronpk-256.jpg'),
    ('/snarfed.org', 'https://snarfed.org/ryan_profile_square_thumb.jpg'),
    ('/ben.thatmustbe.me/?alt=robohash', 'https://ben.thatmustbe.me/image/static/icon_144.jpg'),
    ('/kylewm.com/', 'https://kylewm.com/static/img/users/kyle.jpg'),
    ('/fiatjaf.alhur.es/?alt=robohash', 'http://robohash.org/fiatjaf.alhur.es'),
    ('/cweiske.de?acceptsmall=n', 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||cweiske.de||'),
    ('/cweiske.de', 'http://cweiske.de/favicon.ico'),
    ('/fiatjaf.withknown.com', 'https://fiatjaf.withknown.com/file/5145a8fbe4ad4422adba1bf0e896d2a9/thumb.jpg'),
    ('/http://fiatjaf.withknown.com', 'https://fiatjaf.withknown.com/file/5145a8fbe4ad4422adba1bf0e896d2a9/thumb.jpg'),
    ('/fiatjaf.gmail.com.questo.email', 'https://secure.gravatar.com/avatar/b760f503c84d1bf47322f401066c753f?d=blank&s=200'),
    ('/http://fiatjaf.gmail.com.questo.email', 'https://secure.gravatar.com/avatar/b760f503c84d1bf47322f401066c753f?d=blank&s=200'),
    ('/ben.thatmustbe.me/?d=robohash&f=y', 'http://robohash.org/ben.thatmustbe.me'),
    ('/ben.thatmustbe.me/?d=identicon&f=y', 'https://secure.gravatar.com/avatar/eb8de282b4131c4ab253871ff867c87a?d=identicon'),
    ('/ben.thatmustbe.me/?d=identicon&f=y', 'https://secure.gravatar.com/avatar/eb8de282b4131c4ab253871ff867c87a?d=identicon'),
    ('/https://trello.com/fiatjaf', 'https://trello-avatars.s3.amazonaws.com/d2f9f8c8995019e2d3fda00f45d939b8/170.png'),
    ('/github.com/fiatjaf', 'https://avatars.githubusercontent.com/u/1653275?v=3'),
    ('/https://twitter.com/fiatjaf', 'http://res.cloudinary.com/hafxoat4q/image/twitter_name/w_200/fiatjaf'),
    ('/www.facebook.com/1118751491487611', 'https://graph.facebook.com/v2.2/1118751491487611/picture?type=large'),
    ('/instagram.com/multi', 'http://res.cloudinary.com/hafxoat4q/image/instagram_name/w_200/multi'),
    ('/https://plus.google.com/u/0/109353802488562702914/posts', 'http://res.cloudinary.com/hafxoat4q/image/gplus/109353802488562702914'),
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
    print ''' - {}:
  expected {}
       got {}
    '''.format(url, expected, location)
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
