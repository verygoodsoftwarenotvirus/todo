import { expect } from 'chai';
import { describe } from 'mocha';
import { inheritQueryFilterSearchParams } from './earl';

describe('earl', () => {
  describe('inheritQueryFilterSearchParams', () => {
    it('should be empty for empty input', () => {
      const input = new URLSearchParams();
      const expectation = '';

      expect(inheritQueryFilterSearchParams(input).toString()).to.equal(
        expectation,
      );
    });
  
    it('should ignore non-numeric values', () => {
      const expectedPage = '12345',
        expectedCreatedBefore = 'lol',
        expectedCreatedAfter = 'lol',
        expectedUpdatedBefore = 'lol',
        expectedUpdatedAfter = 'lol',
        expectedIncludeArchived = 'true',
        expectedSortBy = 'asc';
    
      const input = new URLSearchParams([
        ['page', expectedPage],
        ['createdBefore', expectedCreatedBefore],
        ['createdAfter', expectedCreatedAfter],
        ['updatedBefore', expectedUpdatedBefore],
        ['updatedAfter', expectedUpdatedAfter],
        ['includeArchived', expectedIncludeArchived],
        ['sortBy', expectedSortBy],
      ]);
    
      const expectation = `page=${expectedPage}&` +
        `includeArchived=${expectedIncludeArchived}&` +
        `sortBy=${expectedSortBy}`;
    
      expect(inheritQueryFilterSearchParams(input).toString()).to.equal(
        expectation,
      );
    });
    
    it('should be fleshed out with full input', () => {
      const expectedPage = '12345',
        expectedCreatedBefore = '23456',
        expectedCreatedAfter = '34567',
        expectedUpdatedBefore = '45678',
        expectedUpdatedAfter = '56789',
        expectedIncludeArchived = 'true',
        expectedSortBy = 'asc';
    
      const input = new URLSearchParams([
        ['page', expectedPage],
        ['createdBefore', expectedCreatedBefore],
        ['createdAfter', expectedCreatedAfter],
        ['updatedBefore', expectedUpdatedBefore],
        ['updatedAfter', expectedUpdatedAfter],
        ['includeArchived', expectedIncludeArchived],
        ['sortBy', expectedSortBy],
      ]);
    
      const expectation = `page=${expectedPage}&` +
        `createdBefore=${expectedCreatedBefore}&` +
        `createdAfter=${expectedCreatedAfter}&` +
        `updatedBefore=${expectedUpdatedBefore}&` +
        `updatedAfter=${expectedUpdatedAfter}&` +
        `includeArchived=${expectedIncludeArchived}&` +
        `sortBy=${expectedSortBy}`;
    
      expect(inheritQueryFilterSearchParams(input).toString()).to.equal(
        expectation,
      );
    });
  });
});
