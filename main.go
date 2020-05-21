package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/garabais/kpmp/kpage"
)

var maxIter, pages int
var firstLine, input, output string
var timer bool

func main() {
	// Defining flags to parse
	flag.IntVar(&maxIter, "n", 100, "Sets the number of iterations without changes needed to exit")
	flag.IntVar(&pages, "k", 0, "Sets the number of pages in the problem")
	flag.StringVar(&input, "i", "", "Path to the instance of the graph (required)")
	flag.StringVar(&output, "o", "", "Sets the output of the solution, if file alredy exist file is truncated (Default: stdout)")
	flag.BoolVar(&timer, "t", false, "Measure the time of the algoritm(time is displayed in stderr)")

	flag.Parse()

	// Cheching correct values of providen flags
	if input == "" {
		fmt.Println("Provide a instance(-s)")
		flag.PrintDefaults()
		os.Exit(1)
	} else if pages <= 0 {
		fmt.Println("Number of Pages(-k) given is invalid")
		flag.PrintDefaults()
		os.Exit(1)
	} else if maxIter < 0 {
		fmt.Println("Iterations(-n) should be a positive number")
		os.Exit(1)
	}

	// Open input file
	f, err := os.Open(input)
	if err != nil {
		log.Fatalf("Input file: %v", err)
	}
	defer f.Close()

	// Oper output file os stdout
	var out io.Writer
	if output == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(output)
		if err != nil {
			log.Fatalf("Output file: %v", err)
		}
		defer f.Close()
	}

	// Reading first line and storing it
	r := bufio.NewReader(f)
	temp, err := r.ReadBytes('\n')
	if err != nil {
		log.Fatal(err)
	}

	firstLine = string(temp[:len(temp)-1])

	// Start measuring time
	start := time.Now()

	// Solve the problem
	s, err := solve(r, out, pages)
	if err != nil {
		log.Fatal(err)
	}

	// if requiered print exec time
	elapsed := time.Since(start)
	if timer {
		log.Printf("Execution time %s", elapsed)
	}

	// Print solution
	printSolution(out, s)
}

func solve(in io.Reader, out io.Writer, k int) (*kpage.Solution, error) {
	var s *kpage.Solution

	// Generate initial solution
	s, err := generateInitialSolution(in, out, k)
	if err != nil {
		return nil, err
	}

	// Find the local minimum an asign it as the current solution
	s, err = localMinimum(s)
	if err != nil {
		return nil, err
	}

	// Until termination conditions not met pertubate and try to get a better solution
	for i := uint(0); i < uint(maxIter); i++ {
		// Get a copy of the solution with a perturbation
		sp, err := pertubation(s)
		if err != nil {
			return nil, err
		}

		// Find the local minimum of the pertubated solution
		sp, err = localMinimum(sp)
		if err != nil {
			return nil, err
		}

		// ApplyAcceptanceCriterion
		if sp.Crossings < s.Crossings {
			s = sp
			i = 0
		}
	}

	return s, nil
}

func generateInitialSolution(in io.Reader, out io.Writer, k int) (*kpage.Solution, error) {
	var v, e, src, dst uint
	_, err := fmt.Fscanln(in, &v, &v, &e)
	if err != nil {
		return nil, err
	}

	edg := make([]*kpage.Edge, e)

	for i := uint(0); i < e; i++ {

		_, err = fmt.Fscanln(in, &src, &dst)
		if err != nil {
			return nil, err
		}

		edg[i] = kpage.NewEdge(src, dst)
	}

	s, err := kpage.Solve(edg, v, uint(k))

	if err != nil {
		return nil, err
	}

	return s, nil
}

func localMinimum(s *kpage.Solution) (*kpage.Solution, error) {

	// Select a random vertex position
	rand.Seed(time.Now().UTC().UnixNano())
	i := uint(rand.Intn(int(s.Vertex)) + 1)

	// Copy the solution to avoid modify the best solution
	sc, err := s.Copy()

	if err != nil {
		return nil, err
	}

	// Swap with all others vetexes
	for j := uint(1); j <= s.Vertex; j++ {
		if i == j {
			continue
		}

		sc.Swap(i, j)

		// Recalculate the maximum crossings and reasing pages to all the edges
		sc.AssignPages()

		// If the solution is better write it to the best
		if sc.Crossings > s.Crossings {
			temp, err := sc.Copy()

			if err != nil {
				return nil, err
			}

			// Override the pointer to the copy to avoid rewrite all the information
			s = temp
		}

		// Undo the previus change
		sc.Swap(i, j)

	}
	return s, nil
}

func pertubation(s *kpage.Solution) (*kpage.Solution, error) {

	sc, err := s.Copy()

	if err != nil {
		return nil, err
	}

	// Select a random vertex position
	rand.Seed(time.Now().UTC().UnixNano())
	i := uint(rand.Intn(int(s.Vertex-2)) + 2)

	sc.ResetFrom(i)
	// err = sc.OrderVertexes(i)

	if err != nil {
		return nil, err
	}

	sc.AssignPages()

	return sc, nil
}

func printSolution(out io.Writer, s *kpage.Solution) {
	// Print the comment of the file with added information
	fmt.Fprintf(out, "%v - %v Pages - %v Crossings\n", firstLine, s.Pages, s.Crossings)

	// Print the number of vertexes and the number of edges of the graph
	fmt.Fprintf(out, "%v %v %v\n", s.Vertex, s.Vertex, len(s.Edges))

	// Print the order of the vertex
	oString := fmt.Sprintf("%v", s.Order[1:])
	fmt.Fprintf(out, "%v\n", oString[1:len(oString)-1])

	// Print all the edges with the page where it is
	for _, e := range s.Edges {
		fmt.Fprintln(out, e)
	}
}