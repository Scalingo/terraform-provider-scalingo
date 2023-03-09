data "scalingo_stack" "my_chosen_stack" {
  name = "scalingo-22"
}

resource "scalingo_app" "my_test_app" {
  name = "terraform-testapp"

  stack_id = data.scalingo_stack.my_chosen_stack.id
}
