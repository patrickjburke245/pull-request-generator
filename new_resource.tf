resource "aws_rds_cluster" "app1-rds-cluster" {
    cluster_identifier      = "app1-rds-cluster"
    allocated_storage       = 10
    backup_retention_period = 0
    storage_encrypted       = false
    
    tags = {
        environment = "development"
        managed_by  = "terraform"
    }
}