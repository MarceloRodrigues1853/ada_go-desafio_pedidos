import {defineField, defineType} from 'sanity'

export const businessRule = defineType({
  name: 'businessRule',
  title: 'Regra de negócio',
  type: 'document',
  fields: [
    defineField({name: 'title', title: 'Título', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'description', title: 'Descrição', type: 'text', rows: 4, validation: (rule) => rule.required()}),
    defineField({name: 'aggregate', title: 'Agregado', type: 'string'}),
    defineField({name: 'condition', title: 'Condição', type: 'text', rows: 3}),
    defineField({name: 'consequence', title: 'Consequência', type: 'text', rows: 3}),
    defineField({name: 'codeSources', title: 'Fontes no código', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'status', title: 'Status', type: 'string', options: {list: ['current', 'outdated', 'disputed']}, initialValue: 'current'}),
    defineField({name: 'reviewedAt', title: 'Revisada em', type: 'datetime'}),
  ],
})
