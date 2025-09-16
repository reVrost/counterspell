import {
  Drawer as MantineDrawer,
  DrawerProps as MantineDrawerProps,
} from "@mantine/core";

export const Drawer = ({ children, ...props }: MantineDrawerProps) => {
  return (
    <MantineDrawer
      transitionProps={{ duration: 0 }}
      withOverlay={false}
      radius="md"
      position="right"
      offset={32}
      size="lg"
      styles={{
        header: {
          padding: "16px",
          borderBottom: "1px solid var(--mantine-color-gray-3)",
        },
        title: {
          fontWeight: "500",
          color: "var(--mantine-color-gray-6)",
        },
        content: {
          background: "rgba(255, 255, 255, 0.7)",
          backdropFilter: "blur(6px)",
        },
        body: {
          padding: 0,
        },
      }}
      {...props}
    >
      {children}
    </MantineDrawer>
  );
};
