#!/bin/bash
# Helper to create a flame graph with github.com/uber/go-torch.

mkdir -p svg
go test -bench BenchmarkSetupDirs -cpuprofile=cpu.prof
go-torch revolver.test cpu.prof -f svg/0_BenchmarkSetupDirs.svg 
rm revolver.test
rm cpu.prof

mkdir -p svg
go test -bench BenchmarkCreateFile -cpuprofile=cpu.prof
go-torch revolver.test cpu.prof -f svg/1_BenchmarkCreateFile.svg 
rm revolver.test
rm cpu.prof

mkdir -p svg
go test -bench BenchmarkFileCount -cpuprofile=cpu.prof
go-torch revolver.test cpu.prof -f svg/2_BenchmarkFileCount.svg 
rm revolver.test
rm cpu.prof

mkdir -p svg
go test -bench BenchmarkWriteNew -cpuprofile=cpu.prof
go-torch revolver.test cpu.prof -f svg/3_BenchmarkWriteNew.svg 
rm revolver.test
rm cpu.prof
