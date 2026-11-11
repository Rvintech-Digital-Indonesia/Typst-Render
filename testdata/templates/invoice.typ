#set page(paper: "a4", margin: 2cm)
#set text(size: 11pt)

= Invoice

Customer: {{ typstText .Customer.Name }} \\
Address: {{ typstText .Customer.Address }}

#table(
  columns: (1fr, auto),
  [Product], [Price],
  {{ range .Items }}
  [{{ typstText .Name }}], [{{ .Price }}],
  {{ end }}
)

{{ if .Paid }}Paid{{ else }}Payment pending{{ end }}
