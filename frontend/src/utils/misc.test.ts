import { expect } from 'chai';
import { describe } from 'mocha';
import { isNumeric, parseBool } from './misc';

describe('utils', () => {
  describe('isNumeric', () => {
    it('should determine whether something is numeric', () => {
      const expectations = new Map<string, boolean>([
        ['12345', true],
        ['fart', false],
      ]);

      expectations.forEach((expectation: boolean, input: string) => {
        expect(isNumeric(input)).to.equal(expectation);
      });
    });
  });

  describe('parseBool', () => {
    it('should calculate true for expected values', () => {
      const expectations = new Map<string, boolean | null>([
        ['1', true],
        ['t', true],
        ['true', true],
        ['0', false],
        ['f', false],
        ['false', false],
        ['nonsense', null],
      ]);

      expectations.forEach((expectation: boolean | null, input: string) => {
        expect(parseBool(input)).to.equal(expectation);
      });
    });
  });
});
