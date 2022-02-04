package jobsdb

var _ MultiTenantJobsDB = &MultiTenantLegacy{}

type MultiTenantLegacy struct {
	*HandleT
}

func (mj *MultiTenantLegacy) GetAllJobs(customerCount map[string]int, params GetQueryParamsT, maxDSQuerySize int) []*JobT {
	toQuery := customerCount["0"]
	retryList := mj.GetToRetry(GetQueryParamsT{CustomValFilters: params.CustomValFilters, JobCount: toQuery})
	toQuery -= len(retryList)
	waitList := mj.GetWaiting(GetQueryParamsT{CustomValFilters: params.CustomValFilters, JobCount: toQuery}) //Jobs send to waiting state
	toQuery -= len(waitList)
	unprocessedList := mj.GetUnprocessed(GetQueryParamsT{CustomValFilters: params.CustomValFilters, JobCount: toQuery})

	var list []*JobT
	list = append(list, retryList...)
	list = append(list, waitList...)
	list = append(list, unprocessedList...)

	return list
}
