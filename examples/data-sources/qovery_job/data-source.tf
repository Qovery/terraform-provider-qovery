data "qovery_job" "my_job" {
  id = "<job_id>"
}

# Access the job's attributes
# data.qovery_job.my_job.name
# data.qovery_job.my_job.environment_id
# data.qovery_job.my_job.cpu
# data.qovery_job.my_job.memory
