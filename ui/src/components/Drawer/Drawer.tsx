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
      title="Log Details"
      position="right"
      offset={32}
      size="lg"
      styles={{
        content: {
          background: "rgba(255, 255, 255, 0.7)",
          backdropFilter: "blur(6px)",
        },
      }}
      {...props}
    >
      {children}
    </MantineDrawer>
  );
};
