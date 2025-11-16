env "local" {
  dev = "docker+postgres://docker.io/library/postgres:18/dev?search_path=public"
  schemas = ["public"]

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \" \" }}"
    }
  }
}
