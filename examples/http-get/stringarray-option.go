package main

import "strings"

type stringArrayFlag []string

func (v *stringArrayFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v stringArrayFlag) String() string {
	return strings.Join(v, ",")
}
