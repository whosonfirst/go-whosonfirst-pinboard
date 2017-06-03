# go-whosonfirst-pinboard

Go package for Who's On First related tasks using the Pinboard API.

## Important

Too soon. Move along.

## Install

You will need to have both `Go` and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## What is the what?

First of all, you shouldn't rely on any of the naming conventions yet. Not for packages, not for this repo, not for any of the function names.

The thing we're trying to do here is a two-fold:

1. Associate a URL with one or more WOF IDs
2. Archive that URL

To that end we're using Pinboard to bookmark the URL and assigning a whole bunch of `wof:` related tags for lookup and retrieval. We're also using the Wayback Machine API to create a snapshot of the URL and store the Wayback Machine timestamp for that snapshot as a tag.

And basically it all works so that's great. We might just leave it there and get on with other things. Or maybe we will make the bookmarking and archiving components first class interfaces and simply have Pinboard and the Wayback Machine implement them. I don't know yet.

Mostly just understand that packages like `whosonfirst/bookmark` are potentially misleading and packages like `pinboard` or `internetarchive` are incomplete (we don't implement the entirety of the Pinboard API for example) and may eventually move in to their own repos.

## Tools

_Please write me..._

### wof-archive-url

```
./bin/wof-archive-url  -auth-token **** -url http://www.latimes.com/food/dailydish/la-fo-gold-review-sun-nong-dan-20161103-story.html -wofid 1108802103
{
  "URL": "http://www.latimes.com/food/dailydish/la-fo-gold-review-sun-nong-dan-20161103-story.html", 
  "Title": "Where Jonathan Gold goes for spicy comfort food in Koreatown - LA Times", 
  "Tags": [
    "latimes.com", 
    "restaurant", 
    "korean",
    "wof:placetype=venue", 
    "wof:id=1108802103", 
    "archive:dt=20170531002027", 
    "wof:country_id=85633793", 
    "wof:county_id=102086957", 
    "wof:locality_id=85923517", 
    "wof:neighbourhood_id=85886923", 
    "wof:region_id=85688637", 
    "wof:venue_id=1108802103", 
    "wof:continent_id=102191575"
  ]
}
```

### wof-archive-daemon

_Please write me._

```
$> export PINBOARD_AUTH_TOKEN=pinboard:s33kret
$> ./bin/wof-archive-daemon
```

And then:

```
curl -s 'http://localhost:8080?wof_id=588527589&url=https://www.eastbayexpress.com/oakland/boycotters-condemn-924-gilman-st-projects-ethical-backslide/Content?oid=4807011'

{"URL":"https://www.eastbayexpress.com/oakland/boycotters-condemn-924-gilman-st-projects-ethical-backslide/Content?oid=4807011","Title":"Boycotters Accuse 924 Gilman St. Project of Ethical Backslide |\n   \nEast Bay Express","Tags":["eastbayexpress.com","wof:placetype=venue","wof:id=588527589","archive:dt=20170603132449","wk:page=924_Gilman","venue","music","allages","wof:neighbourhood_id=85876237","wof:region_id=85688637","wof:venue_id=588527589","wof:continent_id=102191575","wof:country_id=85633793","wof:county_id=102086959","wof:locality_id=85921915"]}
```

## See also

* https://pinboard.in/api/
