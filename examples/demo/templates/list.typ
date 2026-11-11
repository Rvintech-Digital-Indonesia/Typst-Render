#set page(paper: "a4")

= Array example

#list(
  {{ range . }}
  [{{ typstText . }}],
  {{ end }}
)
