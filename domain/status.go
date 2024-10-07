package domain

import (
	"encoding/json"
	"strconv"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type StatusGraph struct {
	Current string
	Graph   map[string][]string
}

func NewStatusGraph(v string) *StatusGraph {
	// parse value
	if v != "*" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic("value must be integer")
		}

		if i < 0 || i > 20 {
			panic("value must be in range 0..20")
		}
	}

	return &StatusGraph{
		Current: v,
		Graph:   make(map[string][]string),
	}
}

func NewStatusGraphFromJSON(str string) (root *StatusGraph, err error) {
	var rw map[string][]string
	err = json.Unmarshal([]byte(str), &rw)
	if err != nil {
		logrus.Error("Error on unmarshal json: ", err)
		return nil, err
	}

	graph := make(map[string][]string)

	for v, childs := range rw {
		if _, ok := graph[v]; !ok {
			graph[v] = []string{}
		}

		for _, c := range childs {
			if _, ok := graph[c]; !ok {
				graph[c] = []string{}
			}

			graph[v] = append(graph[v], c)
		}
	}

	return &StatusGraph{
		Current: "0",
		Graph:   graph,
	}, nil
}

func NewStatusGraphFromMap(rw map[string][]string) (root *StatusGraph, err error) {
	graph := make(map[string][]string)

	for v, childs := range rw {
		if _, ok := graph[v]; !ok {
			graph[v] = []string{}
		}

		for _, c := range childs {
			if _, ok := graph[c]; !ok {
				graph[c] = []string{}
			}

			graph[v] = append(graph[v], c)
		}
	}

	return &StatusGraph{
		Current: "0",
		Graph:   graph,
	}, nil
}

func (s *StatusGraph) AddRoute(idx, child string) {
	if _, ok := s.Graph[idx]; !ok {
		s.Graph[idx] = []string{}
	}

	s.Graph[idx] = lo.Uniq(append(s.Graph[idx], child))
}

func (s *StatusGraph) RemoveRouteByValue(idx, child string) {
	if _, ok := s.Graph[idx]; !ok {
		logrus.Warnf("route not found. idx: %v child: %v", idx, child)
		return
	}

	s.Graph[idx] = lo.Filter(s.Graph[idx], func(v string, i int) bool {
		return v != child
	})
}

func CheckPathByValue(sg *StatusGraph, current, value string) (bool, []string) {
	if _, ok := sg.Graph[current]; !ok {
		sg.Current = "0"
	}

	if _, ok := sg.Graph[value]; !ok {
		return false, []string{}
	}

	mp := make(map[string]bool)
	routes := []string{sg.Current}

	var fn func(root *StatusGraph, current, value string, mp map[string]bool, routes []string) (bool, []string)
	fn = func(root *StatusGraph, current, value string, mp map[string]bool, routes []string) (bool, []string) {
		if mp[current] {
			return false, routes
		}

		mp[current] = true

		if current == value || current == "*" {
			return true, routes
		}

		if _, ok := root.Graph[current]; !ok {
			return false, routes
		}

		for _, route := range root.Graph[current] {
			routes = append(routes, route)
			if ok, routes := fn(sg, route, value, mp, routes); ok {
				return true, routes
			}
		}
		return false, routes
	}

	return fn(sg, sg.Current, value, mp, routes)
}
