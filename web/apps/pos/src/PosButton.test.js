import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { PosButton } from '@pos/ui'

describe('PosButton', () => {
  it('defaults to an accessible primary button', () => {
    const wrapper = mount(PosButton, { slots: { default: 'Continue' } })
    expect(wrapper.element.tagName).toBe('BUTTON')
    expect(wrapper.attributes('type')).toBe('button')
    expect(wrapper.classes()).toContain('pos-button--primary')
    expect(wrapper.text()).toBe('Continue')
  })

  it('supports a secondary submit action', () => {
    const wrapper = mount(PosButton, { props: { variant: 'secondary', type: 'submit' }, slots: { default: 'Save' } })
    expect(wrapper.attributes('type')).toBe('submit')
    expect(wrapper.classes()).toContain('pos-button--secondary')
  })
})
