package main

import (
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"os"
	"log"
)

type Playlist struct {
	XMLName xml.Name `xml:"playlist"`
	Aliases []Alias `xml:"alias"`
}

type Alias struct {
	XMLName xml.Name `xml:"alias"`
	Group string `xml:"group,attr"`
	Name string `xml:"name,attr"`
	List string `xml:"list,attr"`
	Priority int // set by us
	TGIDs []TGID `xml:"id"`
}

type TGID struct {
	XMLName xml.Name `xml:"id"`
	Type string `xml:"type,attr"`
	Priority int `xml:"priority,attr"`
	Channel string `xml:"channel,attr"`
	Value int `xml:"value,attr"`
	Min int `xml:"min,attr"`
	Max int `xml:"max,attr"`
}

func main() {
	inBytes, _ := ioutil.ReadAll(os.Stdin)

	tgs, err := os.Create("tgs.tsv")
	if err != nil {
		log.Fatal(err)
	}

	defer tgs.Close()

	wl, err := os.Create("wl.tsv")
	if err != nil {
		log.Fatal(err)
	}

	defer wl.Close()

	var playlist Playlist

	err = xml.Unmarshal(inBytes, &playlist)
	if err != nil {
		log.Fatal(err)
	}


	for ia, a := range playlist.Aliases {
		for _, t := range a.TGIDs {
			if t.Priority != 0 {
				playlist.Aliases[ia].Priority = t.Priority
			}
		}
	}

	for _, a := range playlist.Aliases {
		prioField := ""
		if a.Priority != -1 && a.Priority != 0 {
			prioField = fmt.Sprintf("\t%d", a.Priority)
		}

		for _, t := range a.TGIDs {
			switch t.Type {
			case "talkgroupRange":
				if t.Min != 0 && t.Max != 0 {
					if a.Priority != -1 {
						wl.WriteString(fmt.Sprintf("%d\t%d\n", t.Min, t.Max))
					}
					for i := t.Min; i <= t.Max; i++ {
						tgs.WriteString(fmt.Sprintf("%d\t%s%s\n", i, a.Name, prioField))
					}
				} else {
					log.Printf("%s no val, min, or max", a.Name)
				}
			case "talkgroup":
				if t.Value == 0 {
					log.Printf("%s range no val, min, or max", a.Name)
				} else {
					if a.Priority != -1 {
						wl.WriteString(fmt.Sprintf("%d\n", t.Value))
					}
					tgs.WriteString(fmt.Sprintf("%d\t%s%s\n", t.Value, a.Name, prioField))
				}
			case "broadcastChannel":
			case "priority":
			}
		}
	}
}
