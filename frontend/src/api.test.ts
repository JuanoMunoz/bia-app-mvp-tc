import { describe, expect, it, vi, beforeEach } from 'vitest'
import { getCostSummary, getMeter } from './api'

beforeEach(() => {
    vi.unstubAllGlobals()
})

describe('api client', () => {
    it('returns parsed JSON on success', async () => {
        const stub = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ meter_id: 'M-101' }),
        })
        vi.stubGlobal('fetch', stub)
        await expect(getMeter('M-101')).resolves.toEqual({ meter_id: 'M-101' })
        expect(stub).toHaveBeenCalledOnce()
        expect(String(stub.mock.calls[0][0])).toContain('/api/v1/meters/M-101')
    })

    it('throws backend error message on failure', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 404,
            json: () => Promise.resolve({ error: 'meter not found' }),
        }))
        await expect(getMeter('M-999')).rejects.toThrow('meter not found')
    })

    it('builds cost summary query only with defined params', async () => {
        const stub = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ total_cost: 10, currency: 'COP' }),
        })
        vi.stubGlobal('fetch', stub)
        await getCostSummary('2026-09-01T00:00:00Z', '2026-09-14T23:00:00Z')
        const url = String(stub.mock.calls[0][0])
        expect(url).toContain('/api/v1/billing/summary?')
        expect(url).toContain('from=')
        expect(url).toContain('to=')
    })
})
