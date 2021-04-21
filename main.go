package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/RH12503/Triangula/algorithm"
	"github.com/RH12503/Triangula/algorithm/evaluator"
	"github.com/RH12503/Triangula/generator"
	image2 "github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/mutation"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/RH12503/Triangula/render"
	"github.com/RH12503/Triangula/triangulation"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"syscall/js"
)

var algo algorithm.Algorithm
var img image2.Data

func initAlgorithm(this js.Value, i []js.Value) interface{} {
	base64Input := i[0].String()
	points := i[1].Int()

	base64Data := base64Input[strings.IndexByte(base64Input, ',')+1:]
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Data))

	imageFile, _, err := image.Decode(reader)

	if err != nil {
		return nil
	}


	img = image2.ToData(imageFile)

	evaluatorFactory := func(n int) evaluator.Evaluator {
		return evaluator.NewParallel(img, 20, 5, n)
	}
	var mutator mutation.Method
	mutator = mutation.NewGaussianMethod(2/float64(points), 0.3)

	pointFactory := func() normgeom.NormPointGroup {
		return generator.RandomGenerator{}.Generate(points)
	}

	algo = algorithm.NewModifiedGenetic(pointFactory, 400, 5, evaluatorFactory, mutator)
	return nil
}

func stepAlgorithm(this js.Value, i []js.Value) interface{} {
	algo.Step()

	w, h := img.Size()
	triangles := triangulation.Triangulate(algo.Best(), w, h)
	triangleData := render.TrianglesOnImage(triangles, img)
	data := struct {
		Data []render.TriangleData
		Width, Height int
	}{
		Data: triangleData,
		Width:  w,
		Height: h,
	}

	jsonData, _ := json.Marshal(data)

	return string(jsonData)
}

func registerCallbacks() {
	js.Global().Set("initAlgorithm", js.FuncOf(initAlgorithm))
	js.Global().Set("stepAlgorithm", js.FuncOf(stepAlgorithm))
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()

	fmt.Println("WASM Go Initialized")

	<-c
}
