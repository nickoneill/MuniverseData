package main

import (
	"encoding/xml"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Base XmlRoute `xml:"route"`
}

type XmlRoute struct {
	RouteShort	string `xml:"tag,attr"`
	RouteTitle	string `xml:"title,attr"`
	StopList 	[]XmlStop `xml:"stop"`
	Directions	[]XmlDirection `xml:"direction"`
}

type XmlStop struct {
	Tag 	int 	`xml:"tag,attr"`
	StopId 	int		`xml:"stopId,attr"`
	Title 	string	`xml:"title,attr"`
	Lat 	float32	`xml:"lat,attr"`
	Lon 	float32	`xml:"lon,attr"`
}

type XmlDirection struct {
	Title 		string 		`xml:"title,attr"`
	Tag 		string 		`xml:"tag,attr"`
	UseForUI	bool		`xml:"useForUI,attr"`
	StopTagList	[]XmlTag 	`xml:"stop"`
	Name		string  	`xml:"name,attr"`
}

type XmlTag struct {
	Tag 	int	`xml:"tag,attr"`
}

// these are direction tags, not stop tags
type DirectionTags struct {
	InboundTag 	string
	OutboundTag	string
}

// stop tags
type SubwayMaps struct {
	InboundTags		map[int]int
	OutboundTags	map[int]int
}

type LineOrder []string

type MuniverseSchema struct {
	StopList 	[]JsonStop
	LineList 	[]JsonLine
	SubwayList	[]JsonSubway
	BuildDate	int64
}

type JsonLine struct {
	Short 			string
	Title 			string
	IBTag			string
	OBTag			string
	IsMetro			bool
	IsHistoric		bool
	FullDesc		string
	InboundDesc		string
	OutboundDesc	string
	InboundTags 	[]int
	OutboundTags 	[]int
	AllLinesSort	int
}

type JsonStop struct {
	Tag 			int
	StopId 			int
	Title 			string
	IBSubwayOrder 	int
	OBSubwayOrder	int
	Lat 			float32
	Lon 			float32
}

type JsonSubway struct {
	Name		string
	IBStopTag	int
	OBStopTag	int
	Order		int
	AboveGround bool
}

var configFiles = []string{"routeConfig-1","routeConfig-10","routeConfig-108","routeConfig-12",
"routeConfig-14","routeConfig-14L","routeConfig-14X","routeConfig-16X","routeConfig-17",
"routeConfig-18","routeConfig-19","routeConfig-1AX","routeConfig-1BX","routeConfig-2",
"routeConfig-21","routeConfig-22","routeConfig-23","routeConfig-24","routeConfig-27",
"routeConfig-28","routeConfig-28L","routeConfig-29","routeConfig-3","routeConfig-30",
"routeConfig-30X","routeConfig-31","routeConfig-31AX","routeConfig-31BX","routeConfig-33",
"routeConfig-35","routeConfig-36","routeConfig-37","routeConfig-38","routeConfig-38AX",
"routeConfig-38BX","routeConfig-38L","routeConfig-39","routeConfig-41","routeConfig-43",
"routeConfig-44","routeConfig-45","routeConfig-47","routeConfig-48","routeConfig-49","routeConfig-5",
"routeConfig-54","routeConfig-56","routeConfig-59","routeConfig-6","routeConfig-60","routeConfig-61",
"routeConfig-66","routeConfig-67","routeConfig-71","routeConfig-71L","routeConfig-76",
"routeConfig-80X","routeConfig-81X","routeConfig-82X","routeConfig-83X","routeConfig-88",
"routeConfig-8AX","routeConfig-8BX","routeConfig-8X","routeConfig-9","routeConfig-90",
"routeConfig-91","routeConfig-9L","routeConfig-F","routeConfig-J","routeConfig-KOWL",
"routeConfig-K","routeConfig-T","routeConfig-LOWL","routeConfig-L","routeConfig-MOWL","routeConfig-M",
"routeConfig-NOWL","routeConfig-N","routeConfig-NX","routeConfig-TOWL"}
var filteredStopEnds = []string{" arr"," Arr"," Outbound"," OB", " Inbound"}
var directionTagMap = make(map[string]DirectionTags)
var descriptionTagMap = make(map[string]string)
var subwayStops = []JsonSubway{{"West Portal Station",6740,6739,1,false},
							   {"Forest Hill Station",5730,6993,2,false},
							   {"Castro Station",5728,6991,3,false},
							   {"Church Station",5726,6998,4,false},
							   {"Van Ness Station",5419,6996,5,false},
							   {"Civic Center Station",5727,6997,6,false},
							   {"Powell Station",5417,6995,7,false},
							   {"Montgomery Station",5731,6994,8,false},
							   {"Embarcadero Station",6992,7217,9,false},
							   {"Folsom & Embarcadero",4509,4510,10,true},
							   {"Brannan & Embarcadero",4506,7145,11,true},
							   {"2nd & King/Ballpark",5234,5237,12,true},
							   {"4th & King/Caltrain",5239,5240,13,true}}
var lineOrder = []string{"F","59","60","61","J","K","L","M","N","T","NX","1","1AX","1BX","2","3","5","6","8X","8AX","8BX","9","9L","10","12","14","14L","14X","16X","17","18","19","21","22","23","24","27","28","28L","29","30","30X","31","31AX","31BX","33","35","36","37","38","38L","38AX","38BX","39","41","43","44","45","47","48","49","54","56","66","67","71","71L","76","80X","81X","82X","83X","88","108","90","91","J OWL","K OWL","L OWL","M OWL","N OWL","T OWL"}

func main() {
	fmt.Println("okok")

	directionTagMap["F"] = DirectionTags{"F__IBCTRO","F__OBCTRO"}
	directionTagMap["NOWL"] = DirectionTags{"N__OWLIB3","N__OWLOB1"}
	directionTagMap["38"] = DirectionTags{"38_IB1","38_OB2"}

	descriptionTagMap["F"] = "Castro District to Fisherman's Wharf via Downtown and Embarcadero"
	descriptionTagMap["J"] = "Balboa Park to Embarcadero Station via Downtown"
	descriptionTagMap["K"] = "Balboa Park to Embarcadero Station via Downtown"
	// descriptionTagMap["KT"] = "Balboa Park to Sunnydale Ave        via Downtown and Ballpark/Caltrain"
	descriptionTagMap["L"] = "SF Zoo to Embarcadero Station       via Downtown"
	descriptionTagMap["M"] = "Balboa Park to Embarcadero Station via Downtown"
	descriptionTagMap["N"] = "Ocean Beach to Ballpark/Caltrain     via Downtown"
	descriptionTagMap["T"] = "West Portal Station to Sunnydale via Downtown and Ballpark/Caltrain"
	descriptionTagMap["K OWL"] = "Overnight K-line bus service between Balboa Park and Downtown/Ferry Plaza"
	descriptionTagMap["L OWL"] = "Overnight L-line bus service between SF Zoo and Downtown/Ferry Plaza"
	descriptionTagMap["M OWL"] = "Overnight M-line bus service between Balboa Park and Downtown/Ferry Plaza"
	descriptionTagMap["N OWL"] = "Overnight N-line bus service between Ocean Beach and Downtown/Caltrain"
	descriptionTagMap["T OWL"] = "Overnight weekend-only T-line bus service between Mid-market and Sunnydale"
	descriptionTagMap["NX"] = "Outer Sunset to Financial District"
	descriptionTagMap["1"] = "Richmond District to Downtown"
	descriptionTagMap["1AX"] = "Outer Richmond to Downtown"
	descriptionTagMap["1BX"] = "Inner Richmond to Downtown"
	descriptionTagMap["2"] = "Richmond District to Downtown"
	descriptionTagMap["3"] = "Richmond District to Downtown"
	descriptionTagMap["5"] = "Richmond District to Downtown via Hayes Valley and Civic Center (24 hour)"
	descriptionTagMap["6"] = "Inner Sunset to Downtown"
	descriptionTagMap["8X"] = "City College to Downtown and Fisherman's Wharf via Visitacion Valley"
	descriptionTagMap["8AX"] = "Visitacion Valley/Cow Palace to Downtown and Chinatown"
	descriptionTagMap["8BX"] = "City College to Downtown and Fisherman's Wharf via Visitacion Valley"
	descriptionTagMap["9"] = "Visitacion Valley to Downtown"
	descriptionTagMap["9L"] = "Visitacion Valley to Downtown"
	descriptionTagMap["10"] = "General Hospital to Pacific Heights Via Downtown and Chinatown"
	descriptionTagMap["12"] = "Mission District to Russian Hill via Downtown"
	descriptionTagMap["14"] = "Daly City to Ferry Plaza via the Mission District and Downtown (24 hour)"
	descriptionTagMap["14L"] = "Daly City to Ferry Plaza via the Mission Disrict and Downtown"
	descriptionTagMap["14X"] = "Daly City to Ferry Plaza via Downtown"
	descriptionTagMap["16X"] = "Outer Sunset to Downtown"
	descriptionTagMap["17"] = "Community service between Parkmerced and West Portal"
	descriptionTagMap["18"] = "Stonestown to the Outer Richmond"
	descriptionTagMap["19"] = "Hunters Point to Fisherman's Wharf via Civic Center"
	descriptionTagMap["21"] = "Inner Richmond to Downtown via Hayes Valley and Civic Center"
	descriptionTagMap["22"] = "Potrero Hill to the Marina District via the Mission District and Japantown (24 hour)"
	descriptionTagMap["23"] = "SF Zoo to Bayview District via West Portal and Glen Park"
	descriptionTagMap["24"] = "Bayview District to Pacific Heights via the Castro District (24 hour)"
	descriptionTagMap["27"] = "Mission District to Russian Hill via Downtown"
	descriptionTagMap["28"] = "Daly City BART Station to the Marina District"
	descriptionTagMap["28L"] = "Daly City BART Station to the Marina District via the Presidio"
	descriptionTagMap["29"] = "Bayview to the Presidio via the Sunset District"
	descriptionTagMap["30"] = "Marina District to Downtown via Northbeach and Chinatown"
	descriptionTagMap["30X"] = "Marina District to Downtown"
	descriptionTagMap["31"] = "Richmond District to Downtown"
	descriptionTagMap["31AX"] = "Outer Richmond to Downtown"
	descriptionTagMap["31BX"] = "Inner Richmond to Downtown"
	descriptionTagMap["33"] = "Mission District to Pacific Heights via the Castro District and Upper Haight"
	descriptionTagMap["35"] = "Comunnity service between Castro Station and Glen Park via Diamond Heights"
	descriptionTagMap["36"] = "Balboa Park to Midtown Terrace"
	descriptionTagMap["37"] = "Twin Peaks to the Haight"
	descriptionTagMap["38"] = "Richmond District to Downtown (24 hour)"
	descriptionTagMap["38L"] = "Richmond District to Downtown"
	descriptionTagMap["38AX"] = "Outer Richmond to Downtown"
	descriptionTagMap["38BX"] = "Inner Richmond to Downtown"
	descriptionTagMap["39"] = "Fisherman's Wharf to Coit Tower via North Beach"
	descriptionTagMap["41"] = "Marina District to Downtown via Northbeach and Chinatown"
	descriptionTagMap["43"] = "Excelsior to Marina District"
	descriptionTagMap["44"] = "Hunters Point to Inner Richmond via Golden Gate Park"
	descriptionTagMap["45"] = "Marina District to Downtown via Northbeach and Chinatown"
	descriptionTagMap["47"] = "Fisherman's Wharf to Ballpark/Caltrain via Civic Center"
	descriptionTagMap["48"] = "Ocean Beach to Potrero Hill"
	descriptionTagMap["49"] = "Fort Mason to City College via Civic Center"
	descriptionTagMap["52"] = "Excelsior to Forest Hill"
	descriptionTagMap["54"] = "Daly City BART to Hunters Point"
	descriptionTagMap["56"] = "Visitacion Valley community service"
	descriptionTagMap["59"] = "Union Square to Fisherman's Wharf via Chinatown and North Beach"
	descriptionTagMap["60"] = "Union Square to Aquatic Park via Chinatown and Russian Hill"
	descriptionTagMap["61"] = "Ferry Plaza to Van Ness Ave via Chinatown and Financial District"
	descriptionTagMap["66"] = "Outer Sunset/Parkside to Inner Sunset"
	descriptionTagMap["67"] = "Bernal Heights to Mission District"
	descriptionTagMap["71"] = "Ocean Beach to Downtown"
	descriptionTagMap["71L"] = "Ocean Beach to Downtown"
	descriptionTagMap["76"] = "Downtown to the Marin Headlands"
	descriptionTagMap["80X"] = "Caltrain to Downtown (morning peak only)"
	descriptionTagMap["81X"] = "Caltrain to Downtown (morning peak only)"
	descriptionTagMap["82X"] = "Caltrain to Levi Plaza (peak hours only)"
	descriptionTagMap["83X"] = "Caltrain to Civic Center (peak hours only)"
	descriptionTagMap["88"] = "Lake Merced to Balboa Park Station/BART"
	descriptionTagMap["90"] = "Visitacion Valley to Fort Mason Via Civic Center"
	descriptionTagMap["91"] = "West Portal to SF State via Bayview, Downtown, and the Marina"
	descriptionTagMap["108"] = "Treasure Island to Transbay Terminal        (24 hour)"

	var ms MuniverseSchema

	for i := 0; i < len(configFiles); i++ {
		fmt.Printf("reading %v\n",configFiles[i])

		filePath := fmt.Sprintf("%s%s",configFiles[i],".xml")
		data,err := ioutil.ReadFile(filePath)
		
		if err != nil {
			fmt.Printf("error reading routeconfig %v: %v\n",configFiles[i],err)
		}

		var c Config
		err = xml.Unmarshal(data, &c)
		if err != nil {
			fmt.Printf("unmarshal error: %v",err)
		}

		// processing the line data
		line := JsonLine{Short:c.Base.RouteShort,Title:c.Base.RouteTitle}
		if line.Short == "F" || line.Short == "59" || line.Short == "60" || line.Short == "61" {
			line.IsHistoric = true
		} else if line.Short == "J" || line.Short == "K" || line.Short == "T" || line.Short == "L" || line.Short == "M" || line.Short == "N" {
			line.IsMetro = true
		}

		for i,v := range(lineOrder) {
			if v == line.Short {
				line.AllLinesSort = i
			}
		}

		if fullDescription,exists := descriptionTagMap[c.Base.RouteShort]; exists {
			line.FullDesc = fullDescription
		}

		for j := 0; j < len(c.Base.Directions); j++ {
			direct := c.Base.Directions[j]

			if direct.UseForUI {
				fmt.Printf("direction %s has %v stops\n",direct.Name,len(direct.StopTagList))

				if directionTags,exists := directionTagMap[c.Base.RouteShort]; exists {
					fmt.Printf("selecting directions based off tag map\n");

					if direct.Tag == directionTags.InboundTag {
						line.IBTag = direct.Tag
						line.InboundDesc = direct.Title
						for k := 0; k < len(direct.StopTagList); k++ {
							line.InboundTags = append(line.InboundTags, direct.StopTagList[k].Tag)
						}
					} else if direct.Tag == directionTags.OutboundTag {
						line.OBTag = direct.Tag
						line.OutboundDesc = direct.Title
						for k := 0; k < len(direct.StopTagList); k++ {
							line.OutboundTags = append(line.OutboundTags, direct.StopTagList[k].Tag)
						}
					}
				} else {
					if strings.Contains(direct.Name, "Inbound") {
						line.IBTag = direct.Tag
						line.InboundDesc = direct.Title
						for k := 0; k < len(direct.StopTagList); k++ {
							line.InboundTags = append(line.InboundTags, direct.StopTagList[k].Tag)
						}
					} else if strings.Contains(direct.Name, "Outbound") {
						line.OBTag = direct.Tag
						line.OutboundDesc = direct.Title
						for k := 0; k < len(direct.StopTagList); k++ {
							line.OutboundTags = append(line.OutboundTags, direct.StopTagList[k].Tag)
						}
					}
				}
			}
		}

		ms.LineList = append(ms.LineList,line)

		// processing the stop data
		for j := 0; j < len(c.Base.StopList); j++ {
			oldstop := c.Base.StopList[j]

			// check for duplicate stopid
			x := 0
			for y := 0; y < len(ms.StopList); y++ {
				if x == 0 {
					if oldstop.Tag == ms.StopList[y].Tag {
						x = y;
					}
				}
			}

			if x == 0 {
				newstop := JsonStop{oldstop.Tag,oldstop.StopId,oldstop.Title,0,0,oldstop.Lat,oldstop.Lon}
				newstop = filterStop(newstop)

				ms.StopList = append(ms.StopList,newstop)
			} else {
				// fmt.Printf("not adding duplicate stop %v\n",oldstop.Title)
			}
		}
	}

	ms.SubwayList = subwayStops

	ms.BuildDate = time.Now().Unix()

	jsonbytes, err := json.Marshal(ms)
	// jsonbytes, err := json.MarshalIndent(ms,"  ","  ")
	if err != nil {
		fmt.Printf("failed to generate json: %v",err)
	}

	err = ioutil.WriteFile("muniversedata.json", jsonbytes, 0777)
	if err != nil {
		fmt.Printf("error writing muniverse data: %v",err)
	} else {
		fmt.Printf("wrote %v stops and %v lines from %v config files\n",len(ms.StopList),len(ms.LineList),len(configFiles))
	}

}

func filterStop(stop JsonStop) JsonStop {
	for i := 0; i < len(filteredStopEnds); i++ {
		stop.Title = strings.Replace(stop.Title, filteredStopEnds[i], "", 1)
	}
	return stop
}
