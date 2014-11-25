import os
import urlparse
import requests
import tweepy

def twitter(url, query):
    name = urlparse.urlparse(url).path.split('/')[1]
    api = tweepy.API(tweepy.AppAuthHandler(consumer_key=os.environ['TWITTER_KEY'], consumer_secret=os.environ['TWITTER_SECRET']))
    return api.get_user(name).profile_image_url

def facebook(url, query):
    id = query.get('id') or urlparse.urlparse(url).path.split('/')[1]
    pic = requests.get('https://graph.facebook.com/' + id + '/?fields=picture').json()['picture']['data']
    if pic['is_silhouette']:
        return None
    return pic['url']

def instagram(url, query):
    return None

silos = {
  'twitter.com': twitter,
  'www.twitter.com': twitter,
  'facebook.com': facebook,
  'www.facebook.com': facebook,
  'fb.me': facebook,
  'www.fb.me': facebook,
  'instagram.com': instagram,
  'www.instagram.com': instagram,
  'instagr.am': instagram
}

def fetch(domain, url, query={}):
    try:
        return silos[domain](url, query)
    except Exception, e:
        print e
        return None
