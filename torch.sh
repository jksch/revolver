#!/bin/bash
# Helper to create a flame graph with github.com/uber/go-torch.
go test -bench . -cpuprofile=cpu.prof
go-torch revolver.test cpu.prof
rm revolver.test
rm cpu.prof
