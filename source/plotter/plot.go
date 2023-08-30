package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)


// Benchmark struct represents the data for a single run of `benchmark.go`
type Benchmark struct {
	Version  	string	`json:"version"`		// "s" or "p"
	TestSize 	string	`json:"testSize"`		// "xsmall", "small", "medium", "large", "xlarge" 
	Threads  	int		`json:"threads"`		// the number of threads used
	Time	 	float64	`json:"time"`			// elapsed time in seconds for the run
}

const resultsFile = "./benchmark/results.txt"
const imagesPartialPath = "./benchmark/"


// computeAverages computes the average time for each testSize and thread count
// return a map of testSize -> (thread count -> average time) eg: "xsmall" -> (1 -> 0.1), (2 -> 0.2), (4 -> 0.4)
func computeAverages(benchmarks []Benchmark) map[string]map[int]float64 {
    avgMap := make(map[string]map[int]float64)
	lenMap := make(map[string]map[int]int)

    for _, benchmark := range benchmarks {
        // initialize sub-map if it doesn't exist
        if _, ok := avgMap[benchmark.TestSize]; !ok {
            avgMap[benchmark.TestSize] = make(map[int]float64)
			lenMap[benchmark.TestSize] = make(map[int]int)
        }

        // add time to the map, we'll divide it by `repeat` later to get the average
        avgMap[benchmark.TestSize][benchmark.Threads] += benchmark.Time
		lenMap[benchmark.TestSize][benchmark.Threads] += 1
    }

    // divide by `repeat` to get the average
    for testSize, subMap := range avgMap {
        for threads := range subMap {
            subMap[threads] /= float64(lenMap[testSize][threads])
        }
    }

    return avgMap
}

// computeSpeedups computes the speedup for each testSize and thread count
// return a map of testSize -> (thread count -> speedup) eg: "xsmall" -> (2 -> 1.5), (4 -> 1.75)
func computeSpeedups(averages map[string]map[int]float64) map[string]map[int]float64{
	
	speedups := make(map[string]map[int]float64)
	
	// compute speedups
	for testSize, subMap := range averages {
		// get the sequential time
		sequentialTime := subMap[1]

		// compute speedup for each thread count
		for threads, time := range subMap {
			if threads == 1 {
				continue
			}
			if _, ok := speedups[testSize]; !ok {
				speedups[testSize] = make(map[int]float64)
			}
			speedup := sequentialTime / time
			speedups[testSize][threads] = speedup
		}
	}
	return speedups
}


// ParseResults parses the 'results.txt' file and returns a map of Benchmark structs
func ParseResults(pathToResultsFile string) []Benchmark {
	file, _ := os.Open(pathToResultsFile)
	defer file.Close()

	benchmarks := make([]Benchmark, 0)
	decoder := json.NewDecoder(file)

	for {
		var bm Benchmark
		if err := decoder.Decode(&bm); err != nil {
			fmt.Println(err)
			break
		}
		benchmarks = append(benchmarks, bm)
	}
	return benchmarks
}


//=============================================================================
// Plotting methods
//=============================================================================
// Customized tick marks for the Y axis
type CustomYTicks struct{}

// forces plotter to show all valus in Y axis
func (CustomYTicks) Ticks(min, max float64) []plot.Tick {
	var newTicks []plot.Tick
	defaultTicks := plot.DefaultTicks{}
	ticks := defaultTicks.Ticks(min, max)
	for _, t := range ticks {
		t.Label = fmt.Sprintf("%.2f", t.Value)
		newTicks = append(newTicks, t)
	}
	return newTicks
}

// customized tick marks for the X axis
type CustomXTicks struct{
	Threads []int
}
// forces plotter to show all number in X axis for which there are values
func (t CustomXTicks) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	for _, thread := range t.Threads {
		if float64(thread) >= min && float64(thread) <= max {
			ticks = append(ticks, plot.Tick{Value: float64(thread), Label: fmt.Sprintf("%d", thread)})
		}
	}
	return ticks
}


// addAxesPadding increases the range of the axes by a percentage
func addAxesPadding(p *plot.Plot, yPadPercent float64, xPadPercent float64) {
	// add some padding to the borders of the plot
	xmin, xmax := p.X.Min, p.X.Max
	ymin, ymax := p.Y.Min, p.Y.Max

	xpadding := (xmax - xmin) * xPadPercent// 2% of range
	ypadding := (ymax - ymin) * yPadPercent // 2% of range

	p.X.Min = xmin - xpadding
	p.X.Max = xmax + xpadding

	p.Y.Min = ymin - ypadding
	p.Y.Max = ymax + ypadding
}

// formatPlot adds padding to the title and axes, sets grid lines, add custom Y ticks
func formatPlot(p *plot.Plot) {
	// Add space between the title and beginning of the plot
	p.Title.Padding = vg.Points(20)
	p.Title.TextStyle.Font.Size = vg.Points(15)

	// Add space between the axes and the plot
	p.X.Label.Padding = vg.Points(5)
	p.Y.Label.Padding = vg.Points(5)

	// Set grid lines
	gridAll := plotter.NewGrid()
	p.Add(gridAll)

	// Force Y axis to show numbers in every tick
	p.Y.Tick.Marker = CustomYTicks{}
}

// return the keys of a map in ascending order
func sortMapKeys(m map[int]float64) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}


//=============================================================================
// Main
//=============================================================================
func main() {
	args := os.Args
	plot_data := true

	// if -noplot, compute averages and speedups but don't plot
	if len(args) >= 2 {
		if args[2] == "-noplot" {
			plot_data = true
		} else {
			panic("Invalid argument")
		}
	}

	if !plot_data {
		return
	}

	// Parse `results.txt` file, and compute average times and speedups
	benchmarks := ParseResults(resultsFile)

	averagesElapsed := computeAverages(benchmarks)
	speedups := computeSpeedups(averagesElapsed)

	//=============================================================================
	// Plot speedups for each testSize
	//=============================================================================
	
	// colors for the lines for each testSize
	colors := map[string]color.RGBA{
		"xsmall":  	{R: 0, G: 0, B: 255, A: 255},
		"small":   	{R: 0, G: 0, B: 100, A: 255},
		"medium": 	{R: 0, G: 255, B: 0, A: 255},
		"large":    {R: 255, G: 0, B: 0, A: 255},
		"xlarge":   {R: 100, G: 0, B: 0, A: 255},
	}
	

	// this plot contains all speedup lines in one chart
	pAll := plot.New()

	// create one plot for each testSize; populate pAll
	colorIndex := 0
	for testSize, threadsData := range speedups {
		// create a new plot
		p := plot.New()
		
		// set the title and axis labels (obs: new lines and spaces for padding)
		p.Title.Text = fmt.Sprintf("\nSpeedup: testSize = %s", testSize)
		p.X.Label.Text = "Number of Threads \n "
		p.Y.Label.Text = "\nSpeedup"

		formatPlot(p)

		// sort thread counts in ascending order to pass to the graph
		threadNums := sortMapKeys(threadsData)

		// Create the plotter.XYs struct using the sorted keys
		pts := make(plotter.XYs, len(threadNums))
		for i, k := range threadNums {
			pts[i].X = float64(k)
			pts[i].Y = threadsData[k]
		}

		// create a line for the testSize
		line, _ := plotter.NewLine(pts)
			
		// line width and color
		line.LineStyle.Width = vg.Points(1)
		line.LineStyle.Color = colors[testSize]
			
		// create markers
		scatter, _ := plotter.NewScatter(pts)
		scatter.GlyphStyle.Color = colors[testSize]
		scatter.GlyphStyle.Radius = vg.Points(2) 

		// Add the line and markers to the plot
		p.Add(line, scatter)
		pAll.Add(line, scatter)
		pAll.Legend.Add(testSize, line)

		// modify the XY ranges in the plot
		addAxesPadding(p, 0.25, 0.05)

		// force X axis to show all threads values
		p.X.Tick.Marker = CustomXTicks{Threads: threadNums}
		pAll.X.Tick.Marker = CustomXTicks{Threads: threadNums}

		// change the color for the next testSize
		colorIndex++

		// save plot to a PNG file
		if err := p.Save(6*vg.Inch, 6*vg.Inch, fmt.Sprintf("%sspeedup-%s.png", imagesPartialPath, testSize)); err != nil {
			panic(err)
		}
	}

	// final formatting and save the plot with all speedups
	addAxesPadding(pAll, 0.2, 0.05)
	// Set the title and axis labels
	pAll.Title.Text = 	"\nSpeedups"
	pAll.X.Label.Text = "Number of Threads \n "
	pAll.Y.Label.Text = "\nSpeedup"
	pAll.Legend.Top = true
	pAll.Legend.Left = true
	formatPlot(pAll)
	if err := pAll.Save(6*vg.Inch, 6*vg.Inch, fmt.Sprintf("%sspeedup-%s.png", imagesPartialPath, "all")); err != nil {
		panic(err)
	}
}

