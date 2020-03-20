package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save/region"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Chunk struct {
	Level struct {
		Entities []Entity
	}
}

type Entity struct {
	CustomName string
	Pos        []float64
	ID         string `nbt:"id"`
}

func main() {
	world := ""
	args := os.Args[1:]
	if len(args) >= 1 {
		world = args[0]
	} else {
		fmt.Print("Enter the world directory: ")
		fmt.Scan(&world)
	}

	info, err := os.Stat(world)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("World path doesn't resolve to a directory")
	}

	fmt.Println("Searching Overworld")
	searchRegion(world)

	fmt.Println("Searching Nether")
	searchRegion(filepath.Join(world, "DIM-1"))

	fmt.Println("Searching End")
	searchRegion(filepath.Join(world, "DIM1"))

	fmt.Println("Finished")
}

func searchRegion(worldDir string) {
	regionDir := filepath.Join(worldDir, "region")
	regionInfo, err := os.Stat(regionDir)
	if err != nil {
		// soft fail lol
		return
	}

	if !regionInfo.IsDir() {
		panic("Region path doesn't resolve to a directory")
	}

	fmt.Printf(" Directory: %s\n", regionDir)

	files, err := ioutil.ReadDir(regionDir)
	checkErr(err)

	regex := regexp.MustCompile("^r\\.(-?[0-9]+)\\.(-?[0-9]+)\\.mca$")

	regions := make([]string, 0)
	for _, file := range files {
		if regex.MatchString(file.Name()) {
			regions = append(regions, filepath.Join(regionDir, file.Name()))
		}
	}

	fmt.Printf(" Region Files: %d\n", len(regions))
	fmt.Println(" Named Entities:")

	for _, path := range regions {

		r, err := region.Open(path)
		checkErr(err)

		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				if !r.ExistSector(i, j) {
					continue
				}

				data, err := r.ReadSector(i, j)
				checkErr(err)

				decoder, err := read(data)
				checkErr(err)

				chunk := Chunk{}
				err = decoder.Decode(&chunk)
				checkErr(err)

				for _, entity := range chunk.Level.Entities {
					if len(entity.CustomName) > 0 {
						fmt.Printf(
							"  %s (%s) (%.2f, %.2f, %.2f)\n",
							entity.CustomName,
							entity.ID,
							entity.Pos[0],
							entity.Pos[1],
							entity.Pos[2])
					}
				}
			}
		}
		r.Close()
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func read(data []byte) (decoder *nbt.Decoder, err error) {
	var reader io.Reader = bytes.NewReader(data[1:])
	switch data[0] {
	default:
		err = errors.New("unknown compression")
	case 1:
		reader, err = gzip.NewReader(reader)
	case 2:
		reader, err = zlib.NewReader(reader)
	}
	decoder = nbt.NewDecoder(reader)
	return
}
