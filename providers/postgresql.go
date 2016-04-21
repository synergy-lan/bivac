package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

// PostgreSQLProvider is an instance of a Provider interface
// for PostgreSQL backups
type PostgreSQLProvider struct {
	handler   *handler.Conplicity
	vol       *docker.Volume
	backupDir string
}

// GetName returns the provider name
func (p *PostgreSQLProvider) GetName() string {
	return "PostgreSQL"
}

func (p *PostgreSQLProvider) GetHandler() *handler.Conplicity {
	return p.handler
}

func (p *PostgreSQLProvider) GetBackupDir() string {
	return p.backupDir
}

func (p *PostgreSQLProvider) PrepareBackup() (err error) {
	c := p.handler
	vol := p.vol
	log.Infof("Looking for a postgres container using this volume...")
	containers, err := c.ListContainers(docker.ListContainersOptions{})
	checkErr(err, "Failed to list containers: %v", -1)
	for _, container := range containers {
		container, err := c.InspectContainer(container.ID)
		checkErr(err, "Failed to inspect container "+container.ID+": %v", -1)
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.Infof("Volume %v is used by container %v", vol.Name, container.ID)
				log.Infof("Launching pg_dumpall in container %v...", container.ID)
				exec, err := c.CreateExec(
					docker.CreateExecOptions{
						Container: container.ID,
						Cmd: []string{
							"sh",
							"-c",
							"mkdir -p " + mount.Destination + "/backups && pg_dumpall -Upostgres > " + mount.Destination + "/backups/all.sql",
						},
					},
				)

				checkErr(err, "Failed to create exec", 1)

				err = c.StartExec(
					exec.ID,
					docker.StartExecOptions{},
				)

				checkErr(err, "Failed to create exec", 1)

				p.backupDir = "backups"
			}
		}
	}
	return
}