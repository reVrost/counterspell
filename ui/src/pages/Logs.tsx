import { Container, Group, Stack, Title } from "@mantine/core";
import { LogsTable } from "../components/LogsTable";

export default function LogsPage() {
  return (
    <Container size="responsive" px="0">
      {/* <Group */}
      {/*   display="flex" */}
      {/*   h="50px" */}
      {/*   p="xs" */}
      {/*   style={{ */}
      {/*     borderBottom: "1px solid var(--mantine-color-gray-2)", */}
      {/*   }} */}
      {/* ></Group> */}
      <LogsTable />
    </Container>
  );
}
