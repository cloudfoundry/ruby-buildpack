require 'rails_helper'

describe TasksController do
  describe 'GET index.pdf' do
    it 'should be success' do
      create_task_list 1

      get :index, format: :pdf

      expect(response).to be_success
      expect(response.content_type).to eq('application/pdf')
      expect(response.header['Content-Disposition']).to include('tasks.pdf')
    end

    describe 'Responded PDF' do
      before do
        create_task_list 30

        get :index, format: :pdf
      end

      it 'render 30 tasks in list' do
        expect(pdf_texts.grep(/#\d+/).size).to eq(30)
      end

      it 'has 2 pages' do
        expect(pdf_pages.size).to eq(2)
      end
    end
  end

  describe 'GET show/1.pdf' do
    it 'should be success' do
      task = create_task

      get :show, id: task, format: :pdf

      expect(response).to be_success
      expect(response.content_type).to eq('application/pdf')
    end

    describe 'Responded PDF' do
      render_views

      before do
        @task = create_task name: 'Foo'
        get :show, id: @task, format: :pdf
      end

      it 'has a page' do
        expect(pdf_pages.size).to eq(1)
      end

      it 'render the attributes of task correctly' do
        expect(pdf_texts).to include(@task.name, @task.state, @task.due_date.strftime('%Y-%m-%d'))
      end
    end
  end

  def create_task_list(count = 1, attrs = nil)
    count.times { create_task(attrs) }
  end

  def create_task(attrs = nil)
    attrs = { name: 'task', done: false, due_date: Date.today }.merge(attrs || {})
    Task.create attrs
  end

  def pdf_texts
    PDF::Inspector::Text.analyze(response.body).strings
  end

  def pdf_pages
    PDF::Inspector::Page.analyze(response.body).pages
  end
end
