#!/bin/bash
table_name="sgt_service_tree_cloud_instance_type"
mysql_str="mysql -uroot -pmysql123   -b local_test "



function insert_sql() {

    local file_name=$1
    local c_type=$2
    while read instance_type cpu mem; do
        [ -z "$instance_type" ] && continue
        echo "INSERT INTO ${table_name} (cloud_provider,resource_type,instance_type,cpu,mem) VALUES ('${c_type}','ecs','$instance_type', $cpu, $mem);"
    done < ${file_name} |  $mysql_str
    echo "${c_type} file: $file_name add num :`wc -l $file_name`"
}

function drop_create_table() {
cat << EOF | $mysql_str
DROP TABLE IF EXISTS cloud_instance_type;
CREATE TABLE cloud_instance_type (
  id int(11) NOT NULL AUTO_INCREMENT,
  resource_type varchar(255) NOT NULL COMMENT '资源类型: ecs /rds/dcs',
  instance_type varchar(255) NOT NULL COMMENT 'ecs规格',
  cloud_provider varchar(255) NOT NULL DEFAULT '' COMMENT '公有云名称',
  cpu int(10) unsigned NOT NULL COMMENT 'vcpus',
  mem int(10) unsigned NOT NULL COMMENT '单位GB',
  PRIMARY KEY (id)
) ENGINE=InnoDB AUTO_INCREMENT=479 DEFAULT CHARSET=utf8;

EOF
}

drop_create_table
insert_sql "hw_ecs_cpu_mem" "huawei"
insert_sql "aws_ecs_cpu_mem" "aws"

