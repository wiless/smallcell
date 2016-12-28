package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/wiless/cellular/deployment"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/wiless/utils/matplot"
	"github.com/wiless/vlib"

	"github.com/wiless/smallcell"
)

func main() {

	rand.Seed(time.Now().Unix())
	var app smallcell.DeploySystem

	// START OMIT
	app.CellRadius = 350
	app.NUEsPerCell = 500
	app.NCells = 19
	app.TxPower = 26
	app.VertTilt = 15
	app.CarriersGHz = vlib.VectorF{1.8}
	app.ActivatePICO = false
	app.SmallCellPowerDB = 23
	app.NSmallCells = 6
	app.Deploy()
	// END OMIT

	cfd, _ := os.Create("sinrtable.csv")
	// cw := csv.NewWriter(cfd)
	cfd.WriteString("TAG,SINR,DISTANCE,RXNODE,TXNODE")
	for _, m := range app.Metric {
		// var record []string
		p1 := app.Singlecell.Nodes[m.BestRSRPNode].Location
		p2 := app.Singlecell.Nodes[m.RxNodeID].Location
		d := p1.DistanceFrom(p2)
		var tag int
		switch {
		case m.BestSINR < 1.1:
			tag = 0
		case m.BestSINR > 1.1 && m.BestSINR < 6.03:
			tag = 1
		case m.BestSINR >= 6.03:
			tag = 2
		}
		str := fmt.Sprintf("\n%d,%3.2f,%3.2f,%d,%d", tag, m.BestSINR, d, m.RxNodeID, m.BestRSRPNode)
		cfd.WriteString(str)
		// record = append(record, fmt.Sprintf("%d", tag))
		// record = append(record, fmt.Sprintf("%3.2f", m.BestSINR))
		// record = append(record, fmt.Sprintf("%3.2f", d))
		// record = append(record, fmt.Sprint(m.RxNodeID))
		// record = append(record, fmt.Sprint(m.BestRSRPNode))

		// log.Printf("Metric %d , %d %f %d", m.RxNodeID, m.BestSINR, m.BestRSRPNode)
		// cw.Write(record)

	}
	cfd.Close()
	// cw.Flush()

	cdf, _ := os.Create("snrcdf.csv")
	cdf.WriteString("SNR,CDF\n")
	for i := 0; i < len(app.CDFx); i++ {
		cdf.WriteString(fmt.Sprintf("%3.2f,%3.2f\n", app.CDFx[i], app.CDF[i]))
	}
	cdf.Close()

	fd, _ := os.Create("plot.svg")
	CreateNodeMap(app, fd)

}

func CreatePlot(app smallcell.DeploySystem, fd *os.File) {

	lengthM := float64(len(app.CDFx)) /// represents 1000m width=1000px
	width := 1000
	PixelsPerUnit := 10.0 //float64(width) / lengthM

	fig := matplot.Create(fd, width, int(width*3/4), 0, 0)
	fig.PixelsPerUnit = PixelsPerUnit
	// fig.Init(1000, 1000, -500, -500)
	// fmt.Println(" LINE Pixels Per Unit ", fig.PixelsPerUnit, len(app.CDFx))
	// 4x4 m building should be visible as 4*PPU x 4*PPU object
	fig.Canvas.Scale(1)

	fig.DrawGrid(lengthM / 10)
	fig.Canvas.Gend()

	fig.Canvas.Gstyle("stroke:green;fill:blue;fill-opacity:0.5;strokeWidth:1")
	fig.Canvas.Circle(0, 0, int(10), fig.Canvas.RGB(255, 127, 0))
	fig.Canvas.Circle(0, 0, int(10), fig.Canvas.RGB(255, 127, 0))

	fig.Canvas.Gend()

	// fig.LinePlot(app.CDFx, app.CDF)
	fig.Close()
}

func BinIndex(sample float64, Bins int, min, max float64) (bin int, delta float64) {

	delta = (max - min) / float64(Bins-1)
	if sample <= min {
		return 0, delta
	}
	if sample >= max {
		return Bins - 1, delta
	}

	return int(math.Floor((sample - min) / delta)), delta

}

func CreateNodeMap(app smallcell.DeploySystem, fd *os.File) {
	lookup := make(map[int]float64)
	lookup[1] = 7
	lookup[19] = 5
	lookup[7] = 3

	lengthM := (lookup[app.NCells] * 2 * app.CellRadius)
	lengthM = 1000 /// represents 1000m width=1000px
	log.Println("Params ")

	width := 1000
	PixelsPerUnit := 3.0 //float64(width) / lengthM

	PixelsPerUnit = float64(width) / lengthM
	log.Printf("AREA  :%vm, IMAGE : %v px, Resolution Pixels/m : %v", lengthM, width, PixelsPerUnit)
	// width = int(PixelsPerUnit * lengthM)
	fig := matplot.Create(fd, width, width, -width/2, -width/2)
	fig.PixelsPerUnit = PixelsPerUnit
	// fig.Init(1000, 1000, -500, -500)
	// fmt.Println(" Pixels Per Unit ", fig.PixelsPerUnit)
	// 4x4 m building should be visible as 4*PPU x 4*PPU object
	fig.Canvas.Scale(1)
	fig.DrawGrid(lengthM / 10)

	fig.Canvas.Gstyle("stroke:green;fill:blue;fill-opacity:0.05;strokeWidth:1")
	// fig.Canvas.Circle(0, 0, int(app.CellRadius*PixelsPerUnit), fig.Canvas.RGBA(255, 127, 0, .12))
	fig.Canvas.Gend()
	// fig.ScatterC(x, y, 5)

	// fig.Canvas.Translate(width/2, width/2)
	// fig.Canvas.Scale(1)
	// fig.ScatterL(app.Singlecell.GetNodeLocations("UE"), c)
	locations := app.Singlecell.GetNodeLocations("UE")
	ueids := app.Singlecell.GetNodeIDs("UE")
	// log.Println("COLOR ", matplot.Palette)
	// log.Println("Total LOCATIONS ", len(locations))
	// log.Println("UE id ", ueids)
	start := 0
	end := start + 7*app.NUEsPerCell - 1
	_ = end
	colors := []color.Color{color.RGBA{R: 255}, color.RGBA{G: 255}, color.RGBA{B: 255}}
	_ = colors
	// c1, _ := colorful.Hex("#fdffcc")
	// c2, _ := colorful.Hex("#242a42")
	c1 := colorful.Color{1, 0, 0}
	c2 := colorful.Color{0, 0, 1}
	_, _ = c1, c2
	Bins := len(matplot.Palette)
	var SINRmin, SINRmax float64 = -10, 40
	var delta float64
	_, delta = BinIndex(0, Bins, SINRmin, SINRmax)
	_ = delta
	// pal3, _ := colorful.HappyPalette(app.NCells)
	fncmap := func(bin int) color.Color {
		t := float64(bin) / float64(Bins)
		_ = t
		// log.Println(bin, Bins, t, delta, t*delta)
		c := matplot.HeatPalette(bin)
		// c := c2.BlendHsv(c1, t)
		return c
	}
	fn := func(sampleIndex int) color.Color {

		// cellidx := math.Ceil(float64(sampleIndex / app.NUEsPerCell))
		// c := c1.BlendHsv(c2, cellidx/float64(app.NCells-1))
		// c := c1.BlendHsv(c2, cellidx/float64(app.NCells-1))

		// SINR based color
		sinr := app.Metric[ueids[sampleIndex]].BestSINR
		bin, delta := BinIndex(sinr, Bins, SINRmin, SINRmax)

		c := matplot.HeatPalette(bin)
		// log.Println("Fn : sinr, bin, Bins, c, len(matplot.Palette) ", sinr, bin, Bins, c, len(matplot.Palette))
		// c := c1.BlendHsv(c2, float64(bin)/float64(Bins))
		_ = delta
		// t := float64(bin) / float64(Bins)
		// log.Println(sinr, bin, Bins, t, delta, t*delta)
		// c = c1.BlendHsv(c2, t)

		// bsid := app.Metric[ueids[sampleIndex]].BestRSRPNode
		// secid := int(math.Floor(float64(bsid) / float64(app.NCells)))
		// c = colors[secid]
		c = matplot.HeatPalette(bin)
		//
		return c
	}
	fnid := func(sampleIndex int) string {
		return fmt.Sprintf("UE%d", ueids[sampleIndex])
	}

	fnmeta := func(sampleIndex int) string {
		uid := ueids[sampleIndex]
		bsid := app.Metric[uid].BestRSRPNode
		txnode := app.Singlecell.Nodes[bsid]
		rxnode := app.Singlecell.Nodes[uid]

		dist := txnode.Location.DistanceFrom(rxnode.Location)
		return fmt.Sprintf("UE%d, SINR = %3.2f, [%s-%d]=%3.2fm ", ueids[sampleIndex], app.Metric[uid].BestSINR, txnode.Type, bsid, dist)
	}
	// for i := 0; i < app.NCells; i++ {

	// for s := 0; s < 3; s++ {

	// c := color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255)), B: uint8(rand.Intn(255))}
	// c := palette.Plan9[i*3]
	// c := colorful.Color{0.313725, 1, 0}

	// log.Println("Starting .. ", start, end, c)
	UEDENSITY := (25854.0 * .30 * 0.80 / 5.0)

	uniform := deployment.RectangularNPoints(complex(0, 0), 1000, 1000, 0, int(UEDENSITY))
	ulocs := vlib.FromVectorC(uniform, 1)

	hloctions := deployment.HexGrid(19, vlib.Origin3D, app.CellRadius, 30)
	fig.Canvas.Gstyle("stroke:black")

	for k, centre := range hloctions {

		hexedges := deployment.HexVertices(centre.Cmplx(), float64(app.CellRadius), 30)
		hx := vlib.NewVectorI(len(hexedges))
		hy := vlib.NewVectorI(len(hexedges))

		for i, val := range hexedges {

			hx[i] = int(real(val) * PixelsPerUnit)
			hy[i] = int(imag(val) * PixelsPerUnit)
		}

		fig.Canvas.Polygon(hx, hy, fig.Canvas.RGBA(k*10, 0, 0, .1))

	}
	fig.Canvas.Gend()
	fig.ScatterLraw(ulocs, 2, fig.Canvas.RGBA(0, 0, 0, .2))
	// fig.ScatterL(locations[start:end], c)

	fig.ScatterLf(locations, fn, fnid, fnmeta)
	// start = end + 1
	// end = start + app.NUEsPerCell - 1
	// }

	// }
	// c := palette.WebSafe[0]

	fig.ScatterMarksLraw("square", app.Singlecell.GetNodeLocations("BS0"), 20, `style="fill:rgb(0,255,255,1)"`)
	if app.ActivatePICO {
		fig.Canvas.Gstyle("fill:white")
		fig.ScatterMarksLraw("eclipse", app.Singlecell.GetNodeLocations("PICO"), 20, `style="fill:rgb(255,255,255,1)"`)
		fig.Canvas.Gend()
	}
	// fig.Canvas.Gend()
	fig.Canvas.Gend()
	var x0, xN, y int
	y = int(PixelsPerUnit * (lengthM * .45))
	w := int(float64(width) / float64(Bins))
	var offset float64 = -float64(width) / 2.0

	fmt.Println(w, float64(Bins), offset)

	for i := 0; i < Bins; i++ {
		clr := fncmap(i)

		r, g, b, _ := clr.RGBA()

		x := int(float64(i*w) + offset)
		if i == 0 {
			x0 = x
		}
		if i == (Bins - 1) {
			xN = x
		}

		// log.Println("COLOfR ", clr)
		// log.Println(uint8(r), (g), (b))
		// log.Println(int(r), int(g), int(b))
		cstring := fig.Canvas.RGB(int(uint8(r)), int(uint8(g)), int(uint8(b)))
		fig.Canvas.Rect(x, y, w, w, cstring)

	}
	fig.Canvas.Text(x0, y, fmt.Sprintf("%3.2f", SINRmin))
	fig.Canvas.Text(xN, y, fmt.Sprintf("%3.2f", SINRmax))
	fig.Close()
}
