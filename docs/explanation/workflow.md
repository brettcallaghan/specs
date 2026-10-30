# Workflow

The following diagram shows how work moves from a product definition to a
reviewed pull request (PR):

```mermaid
flowchart LR
    P["Product<br/>definition"] --> A["/spec<br/>FRD"]
    P --> H["/spec<br/>Architecture<br/>Blueprints"]
    A --> B{{"Product<br/>review"}}
    H --> I{{"Architecture<br/>review"}}
    B --> C["/spec<br/>Feature<br/>Blueprint"]
    I --> C
    C --> E["/groom-feature-<br/>blueprint"]
    E --> F["Agents<br/>implement"]
    F --> J["Agents<br/>review"]
    J -.-> F
    J --> G{{"Engineer<br/>PR review"}}

    style B fill:#fde68a,stroke:#b45309,color:#1c1917
    style I fill:#fde68a,stroke:#b45309,color:#1c1917
    style G fill:#fde68a,stroke:#b45309,color:#1c1917
    style F stroke-dasharray: 4 4
    style J stroke-dasharray: 4 4
```

Work starts with a **product definition**: a high-level description of the
product that you plan to build. It can list the product's features without
their acceptance criteria. The product definition feeds two parallel tracks:

- **Product:** Product defines features and reviews the acceptance criteria
  in each feature requirements document (FRD).
- **Architecture and Engineering:** Engineering designs the systems,
  containers, and components based on the product definition.

Both tracks feed the **Feature Blueprint**, which maps one FRD onto the
architecture. After that, agents do the work. Grooming turns the blueprint
into Work Orders, agents implement them, and agents review, test, and refine
the result. When the implementation is complete, Engineering reviews the PR.

People review at three points, shaded in the diagram: the FRD, the
architecture, and the finished PR. Everything between the approved Feature
Blueprint and the PR is agent work.

For the steps, see [Plan a feature for agents](../how-to/plan-a-feature.md).
