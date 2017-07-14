/*
Convert text input from LSA extracts into a summary and humanized form
*/

// Example of input format;
/*
<<LSA Extract (Dedupe Info)>>     Note: Each BLK=512 Bytes
|CVDEV#|HDEV# |DEDUP|DRDSTS|DRDJOB|Pointer-1 |Pointer-2 |LSA CTL Flag|Fixed pattern output(BLK)|Compression Reduction(BLK)|        Dedup(BLK)    |    Garbge(BLK)       |     metadata(BLK)    |      Zero Data(BLK)  | Efficiency (BLK)   |  Efficiency ratio  | Warning Level |
|------|------|-----|------|------|----------|----------|------------|-------------------------|--------------------------|----------------------|----------------------|----------------------|----------------------|--------------------|--------------------|---------------|
| 0x12 |0x0005| ON  | 0x02 | 0x01 |0xFFFFFFFF|0xFFFFFFFF| 0x90000000 |   0x00000000071F1FE0    |    0x0000000000000000    |  0x00000002FA0D50A0  |  0x000000026F9FF890  |  0x00000000125D6000  |  0x00000000A0BE1150  |    +0x11FED2940[OK]|        +69%        |    A 167%     |
---snip---
|------|------|-----|------|------|----------|----------|------------|-------------------------|--------------------------|----------------------|----------------------|----------------------|----------------------|--------------------|--------------------|---------------|

 Efficiency Level (BLK) : (Compression Reduction + Dedup + Fixed pattern output + Zero Data) < (Garbge + metadata)
 Efficiency ratio:(Garbge + metadata) / (Compression Reduction + Dedup + Fixed pattern output + Zero Data) * 100
 Warning Level: (Garbge + MetaData) / 3TB * 100; 
 A : Alert > 70%
 W : Warning > 50% & < 70%

DRDSTS
Internal status used for CSV status
Internal status                             ?  CSV_Status    meaning
-----------------------------------------------------------------------------------------------------------------------------
(0x00) LSA_DRSTS_OFF   : [OFF         ]      Disabled      Dedupe/Comp is disabled
(0x01) LSA_DRSTS_ENABL : [Enabling    ]      Enabling      Dedupe/Comp is being enabled
(0x02) LSA_DRSTS_ON    : [ON          ]      Enabled       Dedupe/Comp is enabled
(0x03) LSA_DRSTS_DISA1 : [Disabling1  ]      Rehydrating   Dedupe/Comp is being disabled (Dedupe/Comp is being cancelled)
(0x04) LSA_DRSTS_DISA2 : [Disabling2  ]      Rehydrating   Dedupe/Comp is being disabled (pages are being discarded)
(0x05) LSA_DRSTS_REMOV : [Removing    ]      Removing      DRD vol is being deleted
(0x06) LSA_DRSTS_INCON : [Inconsistent]      Failed        User data on DRD is not guaranteed
-----------------------------------------------------------------------------------------------------------------------------

<<LSA POOL Extract (Dedupe Info)>>      Note: Each BLK=512 Bytes
| POOL# |Fixed pattern output(BLK)|Compression Reduction(BLK)|      Dedupe(BLK)     |    Garbage(BLK)      |     Metadata(BLK)    |    Zero Data(BLK)    |
|-------|-------------------------|--------------------------|----------------------|----------------------|----------------------|----------------------|
|  0x0  |   0x000000000EC0F260    |    0x0000000000000000    |  0x0000000AD73EBDE0  |  0x00000002B244EB40  |  0x0000000030396000  |  0x000000018A91CF90  |
|-------|-------------------------|--------------------------|----------------------|----------------------|----------------------|----------------------|

<< Usage Rate for DIR >>
| DIR#|HDVE#|CVDEV|ALLOC(PAGES)| ALLOC(MB) | Usage(%) |LSMVPG(PAGES)|LSMVPG(MB)|CVDEV TYPE| POST Re-Hydration Size (MB) |
|-----|-----|-----|------------|-----------|----------|-------------|----------|----------|-----------------------------|
| 0010| 0001| 0010|  00003520  |  571200   |   18%    |  00000000   |    0     |    DP    |           571200            |
---snip---
|-----|-----|-----|------------|-----------|----------|-------------|----------|----------|-----------------------------|

CVDEV TYPE             : DP = DP-LDEV  or  LS = Log Structured
ALLOC(PAGES)           : Number of PAGES Allocated in Hex ( 42MB/Page)
ALLOC(MB)              : Allocated Pages Amount in MByte
LSMVPG(PAGES)          : Allocated PAGES in Log Structured CVDEV (LS-CVDEV) => Post Dedupe Pages
LSMVPG(MB)             : Post DeDupe Amount in MBytes
POST Re-Hydration Size : Size of DP-CVDEV after Re-Hydration

Usage %                : ALLOC(MB) / 3TB * 100 ( Provided for both DP-CVDEV and LS-CVDEV)
                         A: Alert Usage > 95% ( Guidance is Relevant Only for LS-CVDEV Area)
                         W: Warning Usage > 85% and < 95% ( Guidance is Relevant Only for LS-CVDEV Area)

--- END -----
*/

package main

import "fmt"
import "os"
import "log"
import "flag"
import "path/filepath"
import "bufio"

import "regexp"
import "strings"
import "time"

import "github.com/tealeg/xlsx"

//type SliceSet map[string][]string

type Table struct {
	index []string
	col map[string][]string
}

func NewTable() *Table {
	return &Table{col: make(map[string][]string)}
}

func (t Table) Add(col string, value string) {
	t.col[col] = append(t.col[col], value)
}

var ( // module globals
	flagVerbose bool
	flagQuiet bool
	flagXlsx bool
	dateFormat=time.RFC1123

	section int
	table Table
)

func init() {
	section=0
}

func parsefile(filePath string) ([]string, error) {

  inputFile, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }
  defer inputFile.Close()	

	scanner := bufio.NewScanner(inputFile)
	var results []string
	for scanner.Scan() {
		if output, add := parser(scanner.Text()); add {
			results = append(results, output)
		}
	}
	return results, nil
}

var ( // Regex constants
	rxBlank = regexp.MustCompile(`^\s*$`)
	rxInstruct = regexp.MustCompile(`^\s*(Efficiency|DRDSTS|CVDEV TYPE|Usage)`)
	rxData = regexp.MustCompile(`^\|CVDEV#\||POOL\#|DIR\#`)
)

func parser(line string) (string, bool) {
	if rxBlank.MatchString(line) {
		section=0
	}
	if rxInstruct.MatchString(line) {
		section=1
	}
	if rxData.MatchString(line) {
		section=2
		fmt.Printf("%q", strings.Split(line, "|"))
	}

	// fmt.Printf("Section=%v\n", section)

	if xlFile != nil {
		fmt.Println("xlFile")
	}


	return line, false
}

func main() {
	flag.BoolVar(&flagVerbose, "v", false, "Prints detailed operations")
	flag.BoolVar(&flagQuiet, "q", false, "No output apart from errors")
	flag.BoolVar(&flagXlsx, "x", false, "Do NOT create xlsx file")
	flag.Parse()

	//items := []string{"."}  // default arguments to use if omitted

	if flag.NArg() == 0 {
		flag.PrintDefaults()
		return
	}

	start := time.Now()
	for _, i := range flag.Args() {
		items, err := filepath.Glob(i)
		if err != nil { log.Fatal(err) }
		for _, j := range items {
			if !flagXlsx {
				outpath := path.filepath.join(j, ".xlsx")
				if outpath <> xlFilename {
					if xlFilename != "" { xlFile.save(xlFilename) }
					xlFilename = outpath
					xlFile = xlsx.NewFile()
				}
			}

			output, err := parsefile(j)
			if err != nil { log.Fatal(err) }
			for _, l := range output {
				fmt.Println(l)
			}
		}
	}



	if !flagQuiet {
		elapsed := time.Since(start)
		fmt.Printf("Processed in %v\n" , elapsed)
	}
}


