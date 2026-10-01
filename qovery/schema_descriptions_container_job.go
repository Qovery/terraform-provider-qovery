package qovery

// Descriptions shared by qovery_container and qovery_job, and by their data sources. A resource
// appends its write-side constraint and its default and limits; a data source uses the base
// sentence alone.

const (
	containerResourceDescription = "Manages a Qovery container: a service that runs a prebuilt image from a container registry in its environment."
	jobResourceDescription       = "Manages a Qovery job: a cron job that runs on a schedule, or a lifecycle job that runs when its environment starts, stops or is deleted."
)

// Image attributes of qovery_container and of the image source of qovery_job.
const (
	containerImageRegistryIDDescription = "ID of the `qovery_container_registry` the image is pulled from."
	containerImageNameDescription       = "Name of the image, without the tag, for example `nginx` or `my-org/my-app`."
	containerImageTagDescription        = "Tag of the image, for example `v1.2.3`."
)

const containerAutoDeployDescription = "Whether Qovery redeploys the container when it receives a new image tag."

const (
	jobMaxDurationSecondsDescription = "Maximum run time of the job, in seconds: a job that runs longer is stopped and marked failed."
	jobMaxNbRestartDescription       = "Number of restarts allowed before the job is marked failed. `0` allows none."
	jobPortDescription               = "Port the health checks probe. It is not exposed outside the cluster."
	jobAutoDeployDescription         = "Whether Qovery redeploys the job on every new commit to its branch, or on every new image tag for an `image` source."
)

// Schedule and source of qovery_job.
const (
	jobScheduleDescription        = "Schedule of the job: a cron schedule, or the environment events that run it."
	jobScheduleWriteNote          = " Set either `cronjob`, or at least one of `on_start`, `on_stop` and `on_delete`."
	jobOnStartDescription         = "Command the lifecycle job runs when the environment starts."
	jobOnStopDescription          = "Command the lifecycle job runs when the environment stops."
	jobOnDeleteDescription        = "Command the lifecycle job runs when the environment is deleted."
	jobLifecycleTypeDescription   = "Type of the lifecycle job. A cron job has none."
	jobLifecycleTypeWriteNote     = " It can only be set at creation: changing it fails at plan time."
	jobLifecycleTypeDefaultNote   = "\n\t- Default: `" + jobLifecycleTypeDefault + "` for a lifecycle job, none for a cron job."
	jobCronJobDescription         = "Cron job settings: the job runs `command` on `schedule`."
	jobCronJobScheduleDescription = "Cron expression of the schedule, for example `*/5 * * * *` for every 5 minutes."
	jobCronJobCommandDescription  = "Command the cron job runs."
	jobIconURIDefaultNote         = "\n\t- Default: `" + jobCronIconURIDefault + "` for a cron job, `" + jobLifecycleIconURIDefault + "` for a lifecycle job."

	jobSourceDescription              = "Source of the job image: an image from a container registry, or a Dockerfile to build."
	jobSourceWriteNote                = " Set exactly one of `image` and `docker`."
	jobSourceImageDescription         = "Prebuilt image from a container registry."
	jobSourceDockerDescription        = "Image Qovery builds from a Dockerfile in a git repository."
	jobDockerfileRawDescription       = "Content of the Dockerfile, for a Dockerfile that is not in the repository."
	jobGitRepositoryDescription       = "Git repository the job is built from."
	jobGitRepositoryBranchDescription = "Branch to build."
	// The schema rejects "" in the configuration (see jobRootPathFromAPI).
	jobGitRepositoryRootPathWriteNote = " Use `/`, not an empty string, for the repository root."
)
