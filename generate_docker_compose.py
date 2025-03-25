import sys

def create_file(name_file, number_of_clients):
    with open(name_file, "w") as file:
        file.write("name: tp0\n")
        file.write("services:\n")
        file.write(server_content() + "\n")
        file.write(clients_content(number_of_clients))
        file.write(network_content())
    
def server_content():
    return f"""\
    server:
        container_name: server
        image: server:latest
        entrypoint: python3 /main.py
        environment:
        - PYTHONUNBUFFERED=1
        networks:
        - testing_net
        volumes:
        - ./server/config.ini:/config.ini  
    """

def clients_content(number_of_clients):
    content = ""
    for i in range(1, number_of_clients+1):
        content += f"""\
    client{i}:
        container_name: client{i}
        image: client:latest
        entrypoint: /client
        environment:
        - CLI_ID={i}
        networks:
        - testing_net
        depends_on:
        - server
        volumes:
        - ./client/config.yaml:/config.yaml 
    """
        content += "\n"
    return content

# entrypoint: significa que cuando se levante el contenedor client{i}, el proceso principal que se va a ejecutar 
# será el ejecutable /client dentro del contenedor.

def network_content():
    return """\
networks:
    testing_net:
        ipam:
            driver: default
            config:
                - subnet: 172.25.125.0/24
    """   

    
def main():
    if len(sys.argv) != 3:
        print("Usage: python3 generate_docker_compose.py <name_file> <number_of_clients>")
        sys.exit(1)
    
    name_file = sys.argv[1]
    if not isinstance(name_file, str) or not name_file.endswith(".yaml"):
        print("The file name is not valid")
        sys.exit(1)
        
    number_of_clients = int(sys.argv[2])
    if not isinstance(number_of_clients, int) or number_of_clients < 0:
        print("The number of clients is not valid")
        sys.exit(1)
        
    create_file(name_file, number_of_clients)
    
if __name__ == "__main__":
    main()