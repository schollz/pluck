package pluck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkParseFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p, _ := New()
		p.Verbose(false)
		p.Load("test/config.toml")
		p.PluckFile("test/test.txt")
	}
}

func TestPluck(t *testing.T) {
	p, err := New()
	p.Verbose(false)
	if err != nil {
		t.Error(err)
	}
	err = p.Load("test/config.toml")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 3, len(p.pluckers))
	assert.Equal(t, "0", p.pluckers[0].config.Name)
	assert.Equal(t, -1, p.pluckers[0].config.Limit)
	assert.Equal(t, "options", p.pluckers[1].config.Name)
	assert.Equal(t, "songs", p.pluckers[2].config.Name)

	p.PluckFile("test/test.txt")
	assert.Equal(t, `{
    "0": "Category Archives: Song of the Day Podcast",
    "options": [
        "2009 Countdown",
        "2010 Countdown",
        "2011 Countdown"
    ],
    "songs": [
        "Juana Molina \u0026#8211; Cosoco",
        "Charms – Siren",
        "Daddy Issues \u0026#8211; Locked Out",
        "Cloud Control \u0026#8211; Rainbow City",
        "Kevin Morby \u0026#8211; Come to Me Now",
        "Les Big Byrd \u0026#8211; Two Man Gang",
        "Thunderpussy – Velvet Noose",
        "Captain, We\u0026#8217;re Sinking \u0026#8211; Trying Year",
        "Mammút – “The Moon Will Never Turn On Me”",
        "Songhoy Blues \u0026#8211; Voter"
    ]
}`, p.ResultJSON())

	p.Load("test/food.toml")
	p.PluckURL("http://www.foodnetwork.com/recipes/food-network-kitchen/15-minute-shrimp-tacos-with-spicy-chipotle-slaw-3676441")
	assert.Equal(t, `15-Minute Shrimp Tacos with Spicy Chipotle Slaw Recipe | Food Network Kitchen | Food Network`, p.Result()["title"])

	p, _ = New()
	p.Add(Config{
		Activators:  []string{"X", "Y"},
		Deactivator: "Z",
	})
	p.PluckString("XaZbYcZd")
	assert.Equal(t, `c`, p.Result()["0"])
}
