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
    ('/fiatjaf.alhur.es/?alt=hash', 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|32|r|000|_|458bdc2027b|7b957c57f3|01990dd64b0|2a41476a264|4e7b647a6|42df4325ecc7'),
    ('/fiatjaf.alhur.es/?alt=robohash', 'http://robohash.org/fiatjaf.alhur.es'),
    ('/cweiske.de', 'http://chart.apis.google.com/chart?chst=d_text_outline&chld=666|42|h|000|_|||cweiske.de||'),
    ('/cweiske.de?acceptsmall=true', 'http://cweiske.de/favicon.ico'),
    ('/http://facebook.com/olavo.decarvalho', 'https://fbcdn-profile-a.akamaihd.net/hprofile-ak-xpf1/v/t1.0-1/c28.46.172.172/s50x50/1743592_10152218847092192_1147232169_n.jpg?oh=74ceb77ca1fd1925674f0264712e243d&oe=5504F36D&__gda__=1425768127_6081fd3d1748181ecd856f2e0af75453'),
    ('/https://twitter.com/nntaleb', 'http://pbs.twimg.com/profile_images/491605295852834817/_JrjVr6K_normal.jpeg'),
    ('/https://plus.google.com/102109597967668633775/about ', ''),
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
