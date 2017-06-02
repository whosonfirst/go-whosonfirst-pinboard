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
