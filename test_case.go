package gunit

import (
	"reflect"
	"testing"

	"github.com/smartystreets/gunit/scan"
)

type testCase struct {
	methodIndex int
	description string
	skipped     bool
	long        bool
	parallel    bool

	setups           []int
	teardowns        []int
	innerFixture     *Fixture
	outerFixtureType reflect.Type
	outerFixture     reflect.Value
	positions        scan.TestCasePositions
}

func newTestCase(methodIndex int, method fixtureMethodInfo, config configuration, positions scan.TestCasePositions) *testCase {
	return &testCase{
		parallel:    config.ParallelTestCases(),
		methodIndex: methodIndex,
		description: method.name,
		skipped:     method.isSkippedTest || config.SkippedTestCases,
		long:        method.isLongTest || config.LongRunningTestCases,
		positions:   positions,
	}
}

func (this *testCase) Prepare(setups, teardowns []int, outerFixtureType reflect.Type) {
	this.setups = setups
	this.teardowns = teardowns
	this.outerFixtureType = outerFixtureType
}

func (this *testCase) Run(t *testing.T) {
	t.Helper()

	if this.skipped {
		t.Run(this.description, this.skip)
	} else if this.long && testing.Short() {
		t.Run(this.description, this.skipLong)
	} else {
		t.Run(this.description, this.run)
	}
}

func (this *testCase) skip(innerT *testing.T) {
	innerT.Skip("\n" + this.positions[innerT.Name()])
}
func (this *testCase) skipLong(innerT *testing.T) {
	innerT.Skipf("Skipped long-running test:\n" + this.positions[innerT.Name()])
}
func (this *testCase) run(innerT *testing.T) {
	innerT.Helper()

	if this.parallel {
		innerT.Parallel()
	}
	this.initializeFixture(innerT)
	defer this.innerFixture.finalize()
	this.runWithSetupAndTeardown()
}
func (this *testCase) initializeFixture(innerT *testing.T) {
	innerT.Log("Test definition:\n" + this.positions[innerT.Name()])
	this.innerFixture = newFixture(innerT, testing.Verbose())
	this.outerFixture = reflect.New(this.outerFixtureType.Elem())
	this.outerFixture.Elem().FieldByName("Fixture").Set(reflect.ValueOf(this.innerFixture))
}

func (this *testCase) runWithSetupAndTeardown() {
	this.runSetups()
	defer this.runTeardowns()
	this.runTest()
}

func (this *testCase) runSetups() {
	for _, setup := range this.setups {
		this.outerFixture.Method(setup).Call(nil)
	}
}

func (this *testCase) runTest() {
	this.outerFixture.Method(this.methodIndex).Call(nil)
}

func (this *testCase) runTeardowns() {
	for _, teardown := range this.teardowns {
		this.outerFixture.Method(teardown).Call(nil)
	}
}
