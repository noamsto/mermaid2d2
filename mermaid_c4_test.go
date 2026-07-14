package mermaid2d2

import (
	"testing"
)

func TestMermaidToD2C4(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "context diagram with person, external system, and relationships",
			in: "C4Context\n" +
				"    title System Context\n" +
				"    Person(customer, \"Customer\", \"A customer\")\n" +
				"    System(banking, \"Internet Banking\")\n" +
				"    System_Ext(email, \"Email System\")\n" +
				"    Rel(customer, banking, \"Uses\")\n" +
				"    Rel(banking, email, \"Sends emails using\", \"SMTP\")\n",
			want: "customer: Customer {shape: person; tooltip: A customer}\n" +
				"banking: Internet Banking\n" +
				"email: Email System {style.stroke-dash: 3}\n" +
				"customer -> banking: Uses\n" +
				"banking -> email: Sends emails using (SMTP)\n",
		},
		{
			name: "container diagram with a boundary, Db/Queue variants, BiRel and Rel_Back",
			in: "C4Container\n" +
				"    System_Boundary(b1, \"Internet Banking\") {\n" +
				"        Container(web, \"Web Application\", \"Java, Spring MVC\")\n" +
				"        ContainerDb(db, \"Database\", \"MySQL\")\n" +
				"        ContainerQueue(queue, \"Message Queue\", \"RabbitMQ\")\n" +
				"    }\n" +
				"    Rel(web, db, \"Reads/Writes\")\n" +
				"    BiRel(web, queue, \"Publishes/Consumes\")\n" +
				"    Rel_Back(queue, db, \"Persists\")\n",
			want: "b1: Internet Banking {\n" +
				"  web: Web Application (Java, Spring MVC)\n" +
				"  db: Database (MySQL) {shape: cylinder}\n" +
				"  queue: Message Queue (RabbitMQ) {shape: queue}\n" +
				"}\n" +
				"b1.web -> b1.db: Reads/Writes\n" +
				"b1.web <-> b1.queue: Publishes/Consumes\n" +
				"b1.db -> b1.queue: Persists\n",
		},
		{
			name: "nested boundaries",
			in: "C4Context\n" +
				"    Enterprise_Boundary(e1, \"Enterprise\") {\n" +
				"        System_Boundary(s1, \"System 1\") {\n" +
				"            System(sys1, \"System 1.1\")\n" +
				"        }\n" +
				"    }\n",
			want: "e1: Enterprise {\n" +
				"  s1: System 1 {\n" +
				"    sys1: System 1.1\n" +
				"  }\n" +
				"}\n",
		},
		{
			name: "deployment diagram with nested Deployment_Node and a leaf Node",
			in: "C4Deployment\n" +
				"    Deployment_Node(aws, \"AWS Cloud\") {\n" +
				"        Node(server, \"Web Server\", \"EC2\")\n" +
				"    }\n",
			want: "aws: AWS Cloud {\n" +
				"  server: Web Server (EC2) {shape: package}\n" +
				"}\n",
		},
		{
			name: "relationship targeting a boundary resolves to its qualified path",
			in: "C4Context\n" +
				"    Person(customer, \"Customer\")\n" +
				"    Enterprise_Boundary(e1, \"Enterprise\") {\n" +
				"        System_Boundary(s1, \"System 1\") {\n" +
				"            System(sys1, \"Sys\")\n" +
				"        }\n" +
				"    }\n" +
				"    Rel(customer, s1, \"Uses\")\n",
			want: "customer: Customer {shape: person}\n" +
				"e1: Enterprise {\n" +
				"  s1: System 1 {\n" +
				"    sys1: Sys\n" +
				"  }\n" +
				"}\n" +
				"customer -> e1.s1: Uses\n",
		},
		{
			name: "labels and descriptions containing square brackets are quoted",
			in: "C4Context\n" +
				"    Person(customer, \"Customer\", \"A customer with [VIP] status\")\n" +
				"    System(sys, \"Sys\")\n" +
				"    Rel(customer, sys, \"Uses [important] feature\")\n",
			want: "customer: Customer {shape: person; tooltip: \"A customer with [VIP] status\"}\n" +
				"sys: Sys\n" +
				"customer -> sys: \"Uses [important] feature\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}
