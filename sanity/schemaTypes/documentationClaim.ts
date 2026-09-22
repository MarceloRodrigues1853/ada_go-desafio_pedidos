import {defineField, defineType} from 'sanity'

export const documentationClaim = defineType({
  name: 'documentationClaim',
  title: 'Afirmação documental',
  type: 'document',
  fields: [
    defineField({name: 'title', title: 'Título', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'claim', title: 'Afirmação', type: 'text', rows: 4, validation: (rule) => rule.required()}),
    defineField({name: 'source', title: 'Fonte documental', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'status', title: 'Status', type: 'string', options: {list: ['current', 'outdated', 'disputed']}, initialValue: 'current'}),
    defineField({name: 'codeEvidence', title: 'Evidência no código', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'conflictsWith', title: 'Conflita com', type: 'array', of: [{type: 'reference', to: [{type: 'documentationClaim'}]}]}),
    defineField({name: 'reviewedAt', title: 'Revisada em', type: 'datetime'}),
  ],
})
