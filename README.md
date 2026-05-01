# Helm Charts

This branch is solely a public repository for the Helm Chart packages of the `court` project. It contains only:

- `index.yaml`
- Helm Chart `.tgz` files

## Update Workflow

1. On the `development` branch, package a new Helm Chart release:

    ```bash
    helm package charts/
    ```

2. Switch to `gh-pages`:

    ```bash
    git switch gh-pages
    ```

3. Update `index.yaml`, merging with the existing index:

    ```bash
    helm repo index . --url https://ghdrope.github.io/court/ --merge index.yaml
    ```

4. Commit and push only the new `.tgz` and updated `index.yaml`:

    ```bash
    git add index.yaml <new_chart>.tgz
    git commit -m "Update Helm Chart to version X.Y.Z"
    git push
    ```