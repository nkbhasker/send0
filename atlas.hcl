data "external_schema" "gorm" {
  program = ["go", "run", ".", "migration", "generate"]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/16/dev"
  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}