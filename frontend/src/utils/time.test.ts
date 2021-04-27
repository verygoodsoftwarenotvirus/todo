import { expect } from 'chai';
import { describe } from 'mocha';
import { renderUnixTime } from './time';

describe('time', () => {
  describe('renderUnixTime', () => {
    it('should return a valid string for a valid timestamp', () => {
      const expectations = new Map<number, string>([
        [1234567890, '17:31 02/13/2009'],
      ]);

      expect(renderUnixTime()).to.equal('never');
      expectations.forEach((expectation: string, input: number) => {
        expect(renderUnixTime(input)).to.equal(expectation);
      });
    });
    it('should return never for an invalid timestamp', () => {
      expect(renderUnixTime()).to.equal('never');
    });
  });
});
