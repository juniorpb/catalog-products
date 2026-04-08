# VTEX

**The Enterprise Digital Commerce Platform**

---

## VTEX | AI Coding Interview

### Take-home assessment

### Submission & Timeline

- **Timeline:** You have 48 hours (two days) from the moment you receive this challenge to submit your solution.
- **Submission Method:** Please host your solution in a public GitHub or GitLab repository and share the link replying to the email that you received this assessment.
- **Guideline Adherence:**
  - Before you begin, please carefully review the Guideline Document provided alongside this challenge.
  - Please, ensure your code structure, naming conventions, and documentation align with the standards described there.

---

### The challenge

#### Catalog Consolidation

A traditional e-commerce company will start operating as a marketplace (a store that sells products from other stores). Marketplaces frequently receive product catalogs from multiple sellers (stores that sell their products on the marketplace). Currently, the store has a product catalog that contains all the items it sells. The store must be capable of receiving products from different sellers and adding them to its catalog. It is important to note that it is common for a product to be sold by several stores. Each seller registers their products, so there may be slight variations in the information for the same item across different stores. Duplicating items is undesirable. However, it is crucial to record which sellers offer each product.

You will receive a SQLite database populated with this store's products. This database has only 2 tables: one product table and one table to link products to sellers.

Implement a catalog consolidation system that will receive a file containing products from different sellers and save new products to the catalog database. In the case of a duplicate product, the system should not insert the item into the product table but should link the existing item to the seller in the appropriate table.

You may make changes to the database if you deem them necessary. The use of Al agents in the solution's implementation is permitted and encouraged.
