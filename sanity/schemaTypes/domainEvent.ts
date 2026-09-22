import {defineField, defineType} from 'sanity'

export const domainEvent = defineType({
  name: 'domainEvent',
  title: 'Evento de domínio',
  type: 'document',
  fields: [
    defineField({name: 'name', title: 'Nome', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'producer', title: 'Produtor', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'consumers', title: 'Consumidores', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'trigger', title: 'Condição de publicação', type: 'text', rows: 3}),
    defineField({name: 'effect', title: 'Efeito', type: 'text', rows: 3}),
    defineField({name: 'payloadFields', title: 'Campos do payload', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'failureHandling', title: 'Retry e DLQ', type: 'text', rows: 3}),
    defineField({name: 'codeSources', title: 'Fontes no código', type: 'array', of: [{type: 'string'}]}),
  ],
  preview: {select: {title: 'name', subtitle: 'producer'}},
})
