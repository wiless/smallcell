package smallcell

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	cell "github.com/wiless/cellular"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/vlib"
)

var sector = `{
	"uid": "",
	"NodeID": 0,
	"FreqHz": 0.4,
	"N": 1,
	"Nodes": 360,
	"Omni": false,
	"MfileName": "output.m",
	"VTiltAngle": 12,
	"HTiltAngle": 0,
	"BeamTilt": 0,
	"DisableBeamTit": false,
	"HoldOn": false,
	"VBeamWidth": 15,
	"HBeamWidth": 65,
	"SLAV": 30,
	"ESpacingVFactor": 0.5,
	"ESpacingHFactor": 0,
	"AASArrayType": 0,
	"CurveWidthInDegree": 30,
	"CurveRadius": 1,
	"GainDb": 10
}`

var secangles = vlib.VectorF{0.0, 120.0, -120.0}

type AppParam struct {
	CellRadius       float64 `endpoints:"d=100.0,desc=Radius of the Cell"`
	NCells           int     `endpoints:"d=1,desc=No of Cells"`
	TxPower          float64 `endpoints:"d=22.0"`
	NUEsPerCell      int     `endpoints:"d=30"`
	ActivatePICO     bool
	SmallCellPowerDB float64
	NSmallCells      int
	VertTilt         float64
}

type DeploySystem struct {
	AppParam `json2html:"true"`
	Metric   map[int]cell.LinkMetric `endpoints:"-"`
	CDF      vlib.VectorF
	CDFx     vlib.VectorF

	CarriersGHz    vlib.VectorF
	Singlecell     deployment.DropSystem `json:"-"`
	defaultAAS     antenna.SettingAAS
	systemAntennas map[int]*antenna.SettingAAS `json:"-"`
}

// var nSectors = 1
// var nCellRadius = 500.0
// var nUEPerCell = 400
// var nCells = 19
// var CarriersGHz = vlib.VectorF{1.8}

func init() {
	log.Println("Initialized apps")
	// d.defaultAAS.SetDefault()
	// vlib.LoadStructure("sector.json", &d.defaultAAS)
}

// func main() {
// 	var dp DeploySystem
// 	dp.CellRadius = 100
// 	dp.NCells = 1
// 	dp.NUEsPerCell = 90
// 	dp.TxPower = 22
// 	dp.Deploy()
// }
func (d *DeploySystem) NodeInfo() map[int]deployment.Node {
	return d.Singlecell.Nodes
}
func (d *DeploySystem) Deploy() {
	d.CarriersGHz = vlib.VectorF{1.8}
	d.defaultAAS.SetDefault()

	json.Unmarshal([]byte(sector), &d.defaultAAS)
	// vlib.LoadStructure("sector.json", &d.defaultAAS)

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	// seedvalue = 0
	rand.Seed(seedvalue)

	var plmodel pathloss.OkumuraHata
	// var plmodel walfisch.WalfischIke
	// var plmodel pathloss.SimplePLModel

	d.DeployLayer(&d.Singlecell)

	// singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	// singlecell.SetAllNodeProperty("UE", "AntennaType", 1)
	d.Singlecell.SetAllNodeProperty("UE", "FreqGHz", d.CarriersGHz)
	d.Singlecell.SetAllNodeProperty("PICO", "FreqGHz", d.CarriersGHz)
	layerBS := []string{"BS0", "BS1", "BS2"}
	// layer2BS := []string{"OBS0", "OBS1", "OBS2"}

	var bsids vlib.VectorI
	d.Singlecell.SetAllNodeProperty("BS0", "Orientation", vlib.VectorF{0, d.VertTilt})
	d.Singlecell.SetAllNodeProperty("BS1", "Orientation", vlib.VectorF{120, d.VertTilt})
	d.Singlecell.SetAllNodeProperty("BS2", "Orientation", vlib.VectorF{-120, d.VertTilt})

	for indx, bs := range layerBS {
		d.Singlecell.SetAllNodeProperty(bs, "Orientation", vlib.VectorF{secangles[indx], d.defaultAAS.VTiltAngle})
		d.Singlecell.SetAllNodeProperty(bs, "FreqGHz", d.CarriersGHz)
		newids := d.Singlecell.GetNodeIDs(bs)
		bsids.AppendAtEnd(newids...)
		fmt.Printf("\n %s : %v", bs, newids)
	}

	d.CreateAntennas(bsids)

	// d.CreateOmniAntennas(bsids)
	d.CreateOmniAntennas(d.Singlecell.GetNodeIDs("PICO"))

	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = 10
	wsystem.FrequencyGHz = d.CarriersGHz[0]

	rxids := d.Singlecell.GetNodeIDs("UE")
	_ = rxids

	d.Metric = make(map[int]cell.LinkMetric)

	// baseCells := vlib.VectorI{0, 1, 2}
	// baseCells = baseCells.Scale(nCells)
	// wsystem.ActiveCells.AppendAtEnd(baseCells.Add(0)...)
	// wsystem.ActiveCells.AppendAtEnd(baseCells.Add(69)...)
	//wsystem.ActiveCells.AppendAtEnd(baseCells.Add(4)...)

	// cell := 2
	// startid := 0 + nUEPerCell*(cell)
	// endid := nUEPerCell * (cell + 1)
	// cell0UE := rxids[startid:endid]
	// log.Printf("\n ************** UEs of Cell %d := %v", cell, cell0UE)

	//only if centre cell metrics to be performed
	// selectedUEs := vlib.NewSegmentI(rxids[0], d.NUEsPerCell)

	// Convetionall ALL UE metrics
	selectedUEs := rxids
	sinr := vlib.NewVectorF(selectedUEs.Size())

	if !d.ActivatePICO {
		// add all bs
		wsystem.ActiveCells = d.Singlecell.GetNodeIDs("BS0", "BS1", "BS2") //, "BS1", "BS2"
		log.Printf("Selected BS only %v", wsystem.ActiveCells)
	} else {
		wsystem.ActiveCells = d.Singlecell.GetNodeIDs("BS0", "BS1", "BS2", "PICO") //, "BS1", "BS2"
		log.Printf("Selected BS,PICO only %v", wsystem.ActiveCells)
	}
	wsystem.ActiveCells.Resize(0)
	log.Println("Activated BS ", wsystem.ActiveCells)
	for indx, rxid := range selectedUEs {

		metric := wsystem.EvaluteLinkMetric(&d.Singlecell, &plmodel, rxid, d.Myfunc)
		d.Metric[rxid] = metric
		sinr[indx] = metric.BestSINR

	}

	matlab := vlib.NewMatlab("resultsplot.m")
	defer matlab.Close()
	matlab.Export("sinr", sinr)
	// fmt.Println(sinr)
	bins := 100
	if bins > sinr.Size() {
		bins = sinr.Size()
	}

	// d.CDFx, d.CDF = vlib.CDF(sinr)
	d.CDFx, d.CDF = vlib.NewCDF(sinr, bins)

	matlab.Export("cdfx", d.CDFx)
	matlab.Export("cdfy", d.CDF)

	matlab.Command("cdfplot(sinr);")
	matlab.Command("figure;plot(cdfx,cdfy,'r');")
	matlab.Close()
	// // Dump UE locations
	{

		fid, _ := os.Create("uelocations.dat")
		ueids := d.Singlecell.GetNodeIDs("UE")
		fmt.Fprintf(fid, "% ID\tX\tY\tZ\tSNR")
		for _, id := range ueids {
			node := d.Singlecell.Nodes[id]

			metric, ok := d.Metric[id]
			var snr float64
			if !ok {
				snr = -100
			} else {
				snr = metric.BestSINR
			}

			fmt.Fprintf(fid, "\n%d\t%f\t%f\t%f\t%3.2f", id, node.Location.X, node.Location.Y, node.Location.Z, snr)

		}
		fid.Close()

	}

	// Dump bs nodelocations
	{
		fid, _ := os.Create("bslocations.dat")

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tPower\tdirection")
		for _, id := range bsids {
			node := d.Singlecell.Nodes[id]
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, node.Location.X, node.Location.Y, node.Location.Z, node.TxPowerDBm, node.Direction)

		}
		fid.Close()

	}

	// Dump antenna nodelocations
	{

		fid, _ := os.Create("antennalocations.dat")
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tHDirection\tHWidth")
		for _, id := range bsids {
			ant := d.Myfunc(id)
			// if id%7 == 0 {
			// 	node.TxPowerDBm = 0
			// } else {
			// 	node.TxPowerDBm = 44
			// }
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, ant.Centre.X, ant.Centre.Y, ant.Centre.Z, ant.HTiltAngle, ant.HBeamWidth)

		}
		fid.Close()

	}
	// vlib.DumpMap2CSV("table400.dat", d.Metric)
	// vlib.SaveStructure(d.Metric, "metric400MHz.json", true)
	fmt.Println("\n")
}

/// Calculate Pathloss
func (d *DeploySystem) DeployLayer(system *deployment.DropSystem) {
	setting := system.GetSetting()

	setting = deployment.NewDropSetting()

	// AreaRadius := CellRadius
	/// Should come from API
	// setting.SetCoverage(deployment.CircularCoverage(AreaRadius))

	/// NodeType should come from API calls
	newnodetype := deployment.NodeType{Name: "BS0", Hmin: 30.0, Hmax: 30.0, Count: d.NCells, TxPowerDBm: d.TxPower}
	newnodetype.Mode = deployment.TransmitOnly
	setting.AddNodeType(newnodetype)

	newnodetype = deployment.NodeType{Name: "BS1", Hmin: 30.0, Hmax: 30.0, Count: d.NCells, TxPowerDBm: d.TxPower}
	newnodetype.Mode = deployment.TransmitOnly
	setting.AddNodeType(newnodetype)

	newnodetype = deployment.NodeType{Name: "BS2", Hmin: 30.0, Hmax: 30.0, Count: d.NCells, TxPowerDBm: d.TxPower}
	newnodetype.Mode = deployment.TransmitOnly
	setting.AddNodeType(newnodetype)
	picoPowerdb := d.SmallCellPowerDB
	if !d.ActivatePICO {
		picoPowerdb = -100
	}
	picotype := deployment.NodeType{Name: "PICO", Hmin: 10.0, Hmax: 10.0, Count: d.NSmallCells, TxPowerDBm: picoPowerdb}
	picotype.Mode = deployment.TransmitOnly
	setting.AddNodeType(picotype)

	/// NodeType should come from API calls
	newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: d.NUEsPerCell * d.NCells}
	newnodetype.Mode = deployment.ReceiveOnly

	setting.AddNodeType(newnodetype)

	// vlib.SaveStructure(setting, "depSettings.json", true)

	system.SetSetting(setting)

	system.Init()

	clocations := deployment.HexGrid(d.NCells, vlib.Origin3D, d.CellRadius, 30)
	system.SetAllNodeLocation("BS0", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS1", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS2", vlib.Location3DtoVecC(clocations))

	// Drop PICO nodes
	log.Println("Dropping PICO Nodes ")
	var dp deployment.DropParameter

	dp.InnerRadius = d.CellRadius * .5
	dp.Type = deployment.Circular
	dp.Radius = d.CellRadius
	dp.NCount = picotype.Count
	loc, err := system.Drop(&dp)
	log.Println("err = ", err, loc)

	loc = deployment.AnnularRingEqPoints(complex(0, 0), d.CellRadius, dp.NCount)
	rotate := vlib.GetEJtheta(30)
	loc = loc.ScaleC(rotate)
	// for i, v := range loc {

	// }
	loc3d := vlib.FromVectorC(loc, picotype.Hmax)
	log.Println("err = ", loc3d)
	log.Println("PICo LOCATIONS ", loc3d)

	system.SetAllNodeLocation3D("PICO", loc3d)

	// Workaround else should come from API calls or Databases
	uelocations := d.LoadUELocations(system)
	system.SetAllNodeLocation("UE", uelocations)

}

func (d *DeploySystem) LoadUELocations(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(d.NCells, vlib.FromCmplx(deployment.ORIGIN), d.CellRadius, 30)
	for indx, bsloc := range hexCenters {
		log.Printf("Deployed for cell %d ", indx)
		ulocation := deployment.HexRandU(bsloc.Cmplx(), d.CellRadius, d.NUEsPerCell, 30)
		// for i, v := range ulocation {
		// 	ulocation[i] = v + bsloc.Cmplx()
		// }
		uelocations = append(uelocations, ulocation...)
	}
	return uelocations

}

func (d *DeploySystem) Myfunc(nodeID int) antenna.SettingAAS {
	// atype := d.Singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	obj, ok := d.systemAntennas[nodeID]
	if !ok {
		log.Printf("No antenna created !! for %d ", nodeID)
		return d.defaultAAS
	} else {

		// fmt.Printf("\nNode %d , Omni= %v, Dirction=%v and center is %v", nodeID, obj.Omni, obj.HTiltAngle, obj.Centre)
		return *obj
	}
}

func (d *DeploySystem) CreateAntennas(bsids vlib.VectorI) {
	if d.systemAntennas == nil {
		d.systemAntennas = make(map[int]*antenna.SettingAAS)
	}

	// omni := antenna.NewAAS()
	// sector := antenna.NewAAS()

	// vlib.LoadStructure("omni.json", omni)
	// vlib.LoadStructure("sector.json", sector)

	for _, i := range bsids {
		d.systemAntennas[i] = antenna.NewAAS()
		json.Unmarshal([]byte(sector), d.systemAntennas[i])
		// vlib.LoadStructure("sector.json", d.systemAntennas[i])
		d.systemAntennas[i].FreqHz = d.CarriersGHz[0] * 1.e9
		// d.systemAntennas[i].HBeamWidth = 65

		d.systemAntennas[i].HTiltAngle = d.Singlecell.Nodes[i].Orientation[0]
		d.systemAntennas[i].VTiltAngle = d.Singlecell.Nodes[i].Orientation[1]

		// if nSectors == 1 {
		// 	d.systemAntennas[i].Omni = true
		// } else {
		// 	d.systemAntennas[i].Omni = false
		// }
		d.systemAntennas[i].CreateElements(d.Singlecell.Nodes[i].Location)
		// fmt.Printf("\nType=%s , BSid=%d : System Antenna : %v", system.Nodes[i].Type, i, d.systemAntennas[i].Centre)

		// hgain := vlib.NewVectorF(360)
		// cnt := 0
		// for d := 0; d < 360; d++ {
		// 	hgain[cnt] = d.systemAntennas[i].ElementDirectionHGain(float64(d))
		// 	// hgain[cnt] = d.systemAntennas[i].ElementEffectiveGain(thetaH, thetaV)
		// 	cnt++
		// }

		// cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
	}
}

func (d *DeploySystem) CreateOmniAntennas(txids vlib.VectorI) {
	if d.systemAntennas == nil {
		d.systemAntennas = make(map[int]*antenna.SettingAAS)
	}

	// sector := antenna.NewAAS()

	// vlib.LoadStructure("omni.json", omni)
	// vlib.LoadStructure("sector.json", sector)

	for _, i := range txids {
		d.systemAntennas[i] = antenna.NewAAS()

		json.Unmarshal(deployment.OMNIantenna, d.systemAntennas[i])
		// vlib.LoadStructure("sector.json", d.systemAntennas[i])
		d.systemAntennas[i].FreqHz = d.CarriersGHz[0] * 1.e9
		// d.systemAntennas[i].HBeamWidth = 65

		// d.systemAntennas[i].HTiltAngle = d.Singlecell.Nodes[i].Orientation[0]
		// d.systemAntennas[i].VTiltAngle = d.Singlecell.Nodes[i].Orientation[1]

		// if nSectors == 1 {
		// 	d.systemAntennas[i].Omni = true
		// } else {
		// 	d.systemAntennas[i].Omni = false
		// }
		d.systemAntennas[i].CreateElements(d.Singlecell.Nodes[i].Location)
		// fmt.Printf("\nType=%s , BSid=%d : System Antenna : %v", system.Nodes[i].Type, i, d.systemAntennas[i].Centre)

		// hgain := vlib.NewVectorF(360)
		// cnt := 0
		// for d := 0; d < 360; d++ {
		// 	hgain[cnt] = d.systemAntennas[i].ElementDirectionHGain(float64(d))
		// 	// hgain[cnt] = d.systemAntennas[i].ElementEffectiveGain(thetaH, thetaV)
		// 	cnt++
		// }

		// cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
	}
}
