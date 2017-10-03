# Examples

## Get headlines from news.google.com

```
$ pluck -a 'role="heading"' -a '>' -d '<' -t -s -u 'https://news.google.com/news/?ned=us&hl=en'
```

## Get latest tweets from Donald Trump

```
$ wget https://twitter.com/search\?f\=tweets\&vertical\=default\&q\=from%3ArealDonaldTrump\&src\=typd -O twitter.html
$ pluck  -a '<p ' -a 'data-aria-label-part=' -a '>' -d '</p>' -l -1 -s -f twitter.html
```

## Read comments from Hacker News page

```
$ pluck -s -t -a 'class="c00"' -a '>' -d '</span' -u https://news.ycombinator.com/item\?id=14911467 | more
```


## Get Rotten Tomatoes movie scores

`rt.toml`:

```toml
[[pluck]]
activators = ["@context","schema.org",'"name":"']
deactivator = '"'
name = "name"

[[pluck]]
activators = ["all-critics-numbers","Reviews Counted:","<span>"]
deactivator = "<"
name = "critic_ratings"

[[pluck]]
activators = ["all-critics-numbers","Average Rating:",">"]
deactivator = "/"
name = "average_critic_rating"

[[pluck]]
activators = ["audience-score","Average Rating:",">"]
deactivator = "/"
name = "average_user_rating"

[[pluck]]
activators = ["User Ratings:",">"]
deactivator = "<"
name = "user_ratings"
```

```bash
$ pluck -s -c rt.toml -u https://www.rottentomatoes.com/m/spider_man_homecoming/
```

Returns:

```json
{
    "average_critic_rating": "7.6",
    "average_user_rating": "4.3",
    "critic_ratings": "276",
    "name": "Spider-Man: Homecoming",
    "user_ratings": "85,131"
}
```
