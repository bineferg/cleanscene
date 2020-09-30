# ![CleanScene](https://github.com/bineferg/cleanscene/blob/master/fixtures/logo.jpg)

[web](https://cleanscene.club) :: [insta](https://www.instagram.com/cleanscene.club) :: [fb](https://www.facebook.com/makeacleanscene)

A climate action collective exploring alternative futures for the dance music community.

It is communal spaces, not flights, that are inextricably linked with the music industry. 
Shared experience, not fast travel, is the lifeblood of our scene.

## RA Top 1000

This project was done in an attempt to bring further awareness to the environmental detriment that our scene creates.

`./cmd/fly` takes an input file of the top 1000 DJ list curated by [RA](https://www.residentadvisor.net/dj.aspx), web crawls through each artists event schedule posted by RA, and calculates the carbon impact of that touring schedule.

### Assumptions
We have made a set of assumptions in order to make a reasonable calculation with the public data available. They are as follows:
1. The the city in which an artist is based is taken first from soundcloud, then from RA, if both sources did not provide a city, then the artist was omitted.
1. The artist will fly in and out of the international airport in or closest to that artists home town.
The artist will fly in and out of the international airport closest to the address of the venue of that specific gig.
1. If an artist has more than one gig within two days of eachother, we assume the artist will fly from one gig to the next.
1. If an artist has more than one gig on the same foreign continent, (that is _not_ the continent on which they are based) within two weeks, we assume the artist will fly from one gig to the next.
1. If an artist travels to Austrailia from the Western* hemesphere from which there is no direct flight, we assume they layover in Dubai. 
1. For all other gigs, the artist will return "home" in between.

*Note*: If an event location or flight route could not be found for a gig, the event was left out.

### Want to know your impact?
If you are an artist on the RA 1000 list and want to know, dispute, or contribute (for the sake of accuracy) your environmental impact as a result of fast gigging please email makeacleanscene@gmail.com with your RA artist link and inquiry.


## The Carbon Calculator

We have created a google spreadsheet enhanced with a script to calculate a carbon offset in eu for that artists gig.
If you are an artist, promoter or an agency and would like to contribute please email makeacleanscene@gmail.com and we can send you a scripted sheet.
