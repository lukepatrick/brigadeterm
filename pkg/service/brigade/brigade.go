package brigade

import (
	"fmt"
	"sort"

	"github.com/Azure/brigade/pkg/storage"

	brigademodel "github.com/slok/brigadeterm/pkg/model/brigade"
)

// Service is the service where all the brigade data will be retrieved
// and prepared with the information that this applications needs.
type Service interface {
	// GetProjectBuilds will get one project.
	GetProject(projectID string) (*brigademodel.Project, error)
	// GetProjectLastBuild will get projects last build.
	GetProjectLastBuild(projectID string) (*brigademodel.Build, error)
	// GetProjects will get all the projects that are on brigade.
	GetProjects() ([]*brigademodel.Project, error)
	// GetBuild will get one build.
	GetBuild(buildID string) (*brigademodel.Build, error)
	// GetProjectBuilds will get all the builds of a project in descendant or ascendant order.
	GetProjectBuilds(project *brigademodel.Project, desc bool) ([]*brigademodel.Build, error)
	// GetBuildJobs will get all the jobs of a build in descendant or ascendant order.
	GetBuildJobs(BuildID string, desc bool) ([]*brigademodel.Job, error)
}

// repository will use kubernetes as repository for the brigade objects.
type service struct {
	client storage.Store
}

// NewService returns a new brigade service.
func NewService(brigadestore storage.Store) Service {
	return &service{
		client: brigadestore,
	}
}

func (s *service) GetProject(projectID string) (*brigademodel.Project, error) {
	prj, err := s.client.GetProject(projectID)

	if err != nil {
		return nil, err
	}
	res := brigademodel.Project(*prj)
	return &res, nil
}

func (s *service) GetProjectLastBuild(projectID string) (*brigademodel.Build, error) {
	prj, err := s.client.GetProject(projectID)

	if err != nil {
		return nil, err
	}

	// Get the available builds.
	builds, err := s.GetProjectBuilds(prj, true)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, fmt.Errorf("no builds available")
	}

	// Get last one.
	lastBuild := brigademodel.Build(*builds[0])
	return &lastBuild, nil
}

// GetProjects satisfies Service interface.
func (s *service) GetProjects() ([]*brigademodel.Project, error) {
	prjs, err := s.client.GetProjects()

	if err != nil {
		return nil, err
	}

	// Sort projects by name.
	sort.Slice(prjs, func(i, j int) bool {
		return prjs[i].Name < prjs[j].Name
	})

	prjList := make([]*brigademodel.Project, len(prjs))
	for i, prj := range prjs {
		p := brigademodel.Project(*prj)
		prjList[i] = &p
	}

	return prjList, nil
}

func (s *service) GetBuild(buildID string) (*brigademodel.Build, error) {
	bld, err := s.client.GetBuild(buildID)

	if err != nil {
		return nil, err
	}
	res := brigademodel.Build(*bld)
	return &res, nil
}

// GetAllProjects satisfies Service interface.
func (s *service) GetProjectBuilds(project *brigademodel.Project, desc bool) ([]*brigademodel.Build, error) {
	pr, err := s.client.GetProject(project.ID)
	if err != nil {
		return []*brigademodel.Build{}, err
	}

	builds, err := s.client.GetProjectBuilds(pr)
	if err != nil {
		return []*brigademodel.Build{}, err
	}

	res := make([]*brigademodel.Build, len(builds))
	for i, build := range builds {
		bl := brigademodel.Build(*build)
		res[i] = &bl
	}

	// Order builds in descending order (last ones first).
	sort.Slice(res, func(i, j int) bool {
		if res[i].Worker == nil || res[j].Worker == nil {
			return false
		}

		if desc {
			return res[i].Worker.StartTime.After(res[j].Worker.StartTime)
		}
		return res[i].Worker.StartTime.Before(res[j].Worker.StartTime)
	})

	return res, nil
}

// GetBuildJobs satisfies Repository interface.
func (s *service) GetBuildJobs(BuildID string, desc bool) ([]*brigademodel.Job, error) {
	bl, err := s.client.GetBuild(BuildID)
	if err != nil {
		return []*brigademodel.Job{}, err
	}

	jobs, err := s.client.GetBuildJobs(bl)
	if err != nil {
		return []*brigademodel.Job{}, err
	}
	res := make([]*brigademodel.Job, len(jobs))
	for i, job := range jobs {
		j := brigademodel.Job(*job)
		res[i] = &j
	}

	// Order jobs in ascending order (first ones first).
	sort.Slice(res, func(i, j int) bool {
		if desc {
			return res[i].StartTime.After(res[j].StartTime)
		}
		return res[i].StartTime.Before(res[j].StartTime)

	})

	return res, nil
}

// func (r *repository) transformJobStatusToBuildStatus(status brigade.JobStatus) brigademodel.BuildStatus {
// 	switch status {
// 	case brigade.JobRunning:
// 		return brigademodel.BuildStatusRunning
// 	case brigade.JobSucceeded:
// 		return brigademodel.BuildStatusSucceeded
// 	case brigade.JobFailed:
// 		return brigademodel.BuildStatusFailed
// 	case brigade.JobUnknown:
// 		return brigademodel.BuildStatusUnknown
// 	case brigade.JobPending:
// 		return brigademodel.BuildStatusPending
// 	}

// 	return brigademodel.BuildStatusUnknown
// }
