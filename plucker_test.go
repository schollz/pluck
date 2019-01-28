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

func BenchmarkParseFileStream(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p, _ := New()
		p.Verbose(false)
		p.Load("test/config.toml")
		p.PluckFile("test/test.txt", true)
	}
}

func TestPluck0(t *testing.T) {
	p, _ := New()
	p.Verbose(false)
	err := p.Load("test/config.toml")
	if err != nil {
		t.Error(err)
	}
	err = p.PluckFile("test/test.txt")
	if err != nil {
		t.Error(err)
	}
}
func TestPluck1(t *testing.T) {
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
}`, p.ResultJSON(true))

	p, _ = New()
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

func TestPluck1Stream(t *testing.T) {
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

	p.PluckFile("test/test.txt", true)
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
}`, p.ResultJSON(true))

}

func TestPluck2(t *testing.T) {
	p, err := New()
	p.Verbose(false)
	if err != nil {
		t.Error(err)
	}
	err = p.Load("test/config2.toml")
	if err != nil {
		t.Error(err)
	}

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
        "Kevin Morby \u0026#8211; Come to Me Now"
    ]
}`, p.ResultJSON(true))
}

func TestPluckSongs(t *testing.T) {
	p, err := New()
	p.Verbose(false)
	if err != nil {
		t.Error(err)
	}
	err = p.Load("test/song.toml")
	if err != nil {
		t.Error(err)
	}

	p.PluckFile("test/song.html")
	assert.Equal(t, `{
    "songs": [
        "/music/The+War+on+Drugs/_/An+Ocean+in+Between+the+Waves",
        "/music/The+War+on+Drugs/_/Suffering",
        "/music/Spoon/_/Inside+Out",
        "/music/Real+Estate/_/Crime"
    ]
}`, p.ResultJSON(true))
}

func TestPluckSkipSection(t *testing.T) {
	p, err := New()
	p.Verbose(false)
	if err != nil {
		t.Error(err)
	}
	p.Add(Config{
		Activators:  []string{"Section 2", "a", "href", `"`},
		Permanent:   1,
		Deactivator: `"`,
	})
	err = p.PluckString(`
<h1>Section 1</h1>
<a href="link1">1</a>
<a href="link2">2</a>
<h1>Section 2</h1>
<a href="link3">3</a>
<a href="link4">4</a>
`)
	assert.Nil(t, err)
	assert.Equal(t, `{
    "0": [
        "link3",
        "link4"
    ]
}`, p.ResultJSON(true))
}

func TestPluckCutSection(t *testing.T) {
	p, err := New()
	p.Verbose(false)
	if err != nil {
		t.Error(err)
	}
	p.Add(Config{
		Activators:  []string{"Section 2", "a", "href", `"`},
		Permanent:   1,
		Deactivator: `"`,
		Finisher:    "Section 3",
		Maximum:     6,
	})
	err = p.PluckString(`<h1>Section 1</h1>
<a href="link1">1</a>
<a href="link2">2</a>
<h1>Section 2</h1>
<a href="link3">3</a>
<a href="link4 but this link is too long">4</a>
<h1>Section 3</h1>
<a href="link5">5</a>
<a href="link6">6</a>`)
	assert.Nil(t, err)
	assert.Equal(t, `{"0":"link3"}`, p.ResultJSON())

	assert.Equal(t, []Config{Config{
		Activators:  []string{"Section 2", "a", "href", `"`},
		Permanent:   1,
		Deactivator: `"`,
		Finisher:    "Section 3",
		Limit:       -1,
		Name:        "0",
		Maximum:     6,
	}}, p.Configuration())
}
