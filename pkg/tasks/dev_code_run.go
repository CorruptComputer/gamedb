package tasks

type DevCodeRun struct {
}

func (c DevCodeRun) ID() string {
	return "run-dev-code"
}

func (c DevCodeRun) Name() string {
	return "Run dev code"
}

func (c DevCodeRun) Cron() string {
	return ""
}

func (c DevCodeRun) work() {

}
