# SENG468
UVic course project for SENG 468

This project requires docker installation.
To start the application: 

For MacOSX & Linux systems:  `make build` \
For Windows: `docker compose up --build`

Then, to execute the commands in the userworkload file:
`go run cli.go <relative path_to_user_workload>`

There are several sample userworkload files in the folder called 'user_workload_files'

A log file called 'logfile.xml' will be generated.

To delete all containers and volumes:
For MacOSX & Linux systems: `make clean`\
For Windows: `docker system prune -a`