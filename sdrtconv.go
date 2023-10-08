package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Playlist struct {
	XMLName xml.Name `xml:"playlist"`
	Aliases []Alias  `xml:"alias"`
}

type Alias struct {
	XMLName  xml.Name `xml:"alias"`
	Group    string   `xml:"group,attr"`
	Name     string   `xml:"name,attr"`
	List     string   `xml:"list,attr"`
	Priority int      // set by us
	TGIDs    []TGID   `xml:"id"`
}

type TGID struct {
	XMLName  xml.Name `xml:"id"`
	Type     string   `xml:"type,attr"`
	Priority int      `xml:"priority,attr"`
	Channel  string   `xml:"channel,attr"`
	Value    int      `xml:"value,attr"`
	Min      int      `xml:"min,attr"`
	Max      int      `xml:"max,attr"`
}

func main() {
	var talkgroup, whitelist, rdioCsv *string
	var tsvMode, csvMode *bool
	talkgroup = flag.String("tgs", "tgs.tsv", "tgs.tsv out filename")
	whitelist = flag.String("wl", "wl.tsv", "wl.tsv out filename")
	rdioCsv = flag.String("rdio", "", "rdio.csv out filename")
	tsvMode = flag.Bool("tsv", false, "enable tsv mode")
	csvMode = flag.Bool("csv", false, "enable csv mode")

	flag.Parse()

	playlistXml := flag.Arg(0)
	if playlistXml == "" {
		playlistXml = "default.xml"
	}

	if !*tsvMode && !*csvMode {
		log.Fatal("must specify a mode")
	}

	inBytes, err := os.ReadFile(playlistXml)
	if err != nil {
		log.Fatal(err)
	}

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

	if *tsvMode {
		tgs, err := os.Create(*talkgroup)
		if err != nil {
			log.Fatal(err)
		}

		defer tgs.Close()

		wl, err := os.Create(*whitelist)
		if err != nil {
			log.Fatal(err)
		}

		defer wl.Close()

		err = playlist.GenerateTSV(tgs, wl)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *csvMode {
		var iow io.Writer
		if *rdioCsv == "" {
			iow = os.Stdout
		} else {
			c, err := os.Create(*rdioCsv)
			if err != nil {
				log.Fatal(err)
			}

			defer c.Close()

			iow = c
		}
		err := playlist.GenerateCSV(iow)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (a *Alias) writeCSVTG(tgid int, w io.Writer) error {
	hexTg := strings.Replace(fmt.Sprintf("%X", tgid), "0X", "", 1)
	_, err := io.WriteString(w, fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%d\n",
		tgid,       // dec
		hexTg,      // hex
		"D",        // mode
		a.Name,     // alpha
		a.Name,     // desc
		a.Group,    // tag
		a.List,     // tag
		a.Priority, // prio
	))

	return err
}

// Decimal,Hex,Mode,Alpha Tag,Description,Tag,Tag,Priority,Stream List
func (p *Playlist) GenerateCSV(csv io.Writer) error {
	for _, a := range p.Aliases {
		for _, t := range a.TGIDs {
			switch t.Type {
			case "talkgroupRange":
				if t.Min != 0 && t.Max != 0 {
					for i := t.Min; i <= t.Max; i++ {
						err := a.writeCSVTG(i, csv)
						if err != nil {
							return err
						}
					}
				} else {
					log.Printf("%s no val, min, or max", a.Name)
				}
			case "talkgroup":
				if t.Value == 0 {
					log.Printf("%s range no val, min, or max", a.Name)
				} else {
					err := a.writeCSVTG(t.Value, csv)
					if err != nil {
						return err
					}
				}
			case "broadcastChannel":
			case "priority":
			}
		}
	}
	return nil
}

func (p *Playlist) GenerateTSV(tgs, wl io.Writer) error {
	for _, a := range p.Aliases {
		prioField := ""
		if a.Priority != -1 && a.Priority != 0 {
			prioField = fmt.Sprintf("\t%d", a.Priority)
		}

		for _, t := range a.TGIDs {
			switch t.Type {
			case "talkgroupRange":
				if t.Min != 0 && t.Max != 0 {
					if a.Priority != -1 {
						_, err := io.WriteString(wl, fmt.Sprintf("%d\t%d\n", t.Min, t.Max))
						if err != nil {
							return err
						}
					}
					for i := t.Min; i <= t.Max; i++ {
						_, err := io.WriteString(tgs, fmt.Sprintf("%d\t%s%s\n", i, a.Name, prioField))
						if err != nil {
							return err
						}
					}
				} else {
					log.Printf("%s no val, min, or max", a.Name)
				}
			case "talkgroup":
				if t.Value == 0 {
					log.Printf("%s range no val, min, or max", a.Name)
				} else {
					if a.Priority != -1 {
						_, err := io.WriteString(wl, fmt.Sprintf("%d\n", t.Value))
						if err != nil {
							return err
						}
					}
					_, err := io.WriteString(tgs, fmt.Sprintf("%d\t%s%s\n", t.Value, a.Name, prioField))
					if err != nil {
						return err
					}
				}
			case "broadcastChannel":
			case "priority":
			}
		}
	}

	return nil
}
