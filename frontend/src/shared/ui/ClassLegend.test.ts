// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClassLegend from './ClassLegend.vue'

describe('ClassLegend', () => {
  it('shows all classes by default', () => {
    const wrapper = mount(ClassLegend)

    expect(wrapper.text()).toContain('Рекомендованные')
    expect(wrapper.text()).toContain('Разрешённые')
    expect(wrapper.text()).toContain('Запрещённые')
  })

  it('can omit the forbidden class on the changes page', () => {
    const wrapper = mount(ClassLegend, { props: { includeForbidden: false } })

    expect(wrapper.text()).not.toContain('Запрещённые')
  })
})
