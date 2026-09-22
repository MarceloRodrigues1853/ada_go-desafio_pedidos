import {defineField, defineType} from 'sanity'

export const apiEndpoint = defineType({
  name: 'apiEndpoint',
  title: 'Endpoint da API',
  type: 'document',
  fields: [
    defineField({name: 'title', title: 'Título', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'method', title: 'Método', type: 'string', options: {list: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']}, validation: (rule) => rule.required()}),
    defineField({name: 'path', title: 'Caminho', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'useCase', title: 'Caso de uso', type: 'text', rows: 3}),
    defineField({name: 'requestExample', title: 'Exemplo de request', type: 'text', rows: 5}),
    defineField({name: 'responses', title: 'Respostas', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'relatedRules', title: 'Regras relacionadas', type: 'array', of: [{type: 'reference', to: [{type: 'businessRule'}]}]}),
    defineField({name: 'codeSources', title: 'Fontes no código', type: 'array', of: [{type: 'string'}]}),
  ],
  preview: {select: {title: 'title', method: 'method', path: 'path'}, prepare: ({title, method, path}) => ({title, subtitle: `${method} ${path}`})},
})
